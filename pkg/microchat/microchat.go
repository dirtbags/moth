package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// Send something to the client at least this often, no matter what
const Keepalive = 30 * time.Second

// UserResolver can turn event ID and user ID into a username
type UserResolver interface {
	// Resolve takes an event ID and user ID, and returns a username
	Resolve(string, string) (string, error)
}

// resolver is the UserResolver currently in use for this server instance
var resolver UserResolver

// throttler is our global Throttler
var throttler *Throttler

var rdb *redis.Client

func forumKey(event string, forum string) string {
	return fmt.Sprintf("%s|%s", event, forum)
}

type LogEvent struct {
	Event    string
	User     string
	Username string
	Forum    string
	Text     string
}

func sayHandler(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	user := r.FormValue("user")
	forum := r.FormValue("forum") // this can be empty
	text := r.FormValue("text")

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if (event == "") || (user == "") || (text == "") {
		http.Error(w, "Insufficient Arguments", http.StatusBadRequest)
		return
	}
	if len(text) > 4096 {
		http.Error(w, "Too Long", http.StatusRequestEntityTooLarge)
		return
	}
	logEvent := LogEvent{
		Event: event,
		User:  user,
		Forum: forum,
		Text:  text,
	}

	if username, err := resolver.Resolve(event, user); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		log.Println("Rejected say", event, user, text)
		return
	} else {
		logEvent.Username = username
	}

	if !throttler.CanPost(event, user) {
		log.Println("Rejected (too fast)", logEvent)
		http.Error(w, "Slow Down", http.StatusTooManyRequests)
		return
	}

	rdb.XAdd(
		context.Background(),
		&redis.XAddArgs{
			Stream: forumKey(event, forum),
			ID:     "*",
			Values: map[string]interface{}{
				"user":   user,
				"text":   text,
				"client": r.RemoteAddr,
			},
		},
	)
	log.Println("Posted", logEvent)

	w.WriteHeader(http.StatusOK)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	event := r.FormValue("event")
	user := r.FormValue("user")
	since := r.FormValue("since")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if (event == "") || (user == "") {
		http.Error(w, "Insufficient Arguments", http.StatusBadRequest)
		return
	}

	var fora []string
	for _, forum := range r.Form["forum"] {
		fora = append(fora, forumKey(event, forum))
	}
	if since == "" {
		since = "0"
	}

	if _, err := resolver.Resolve(event, user); err != nil {
		log.Println("Rejected read", event, user)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Cannot flush this connection", http.StatusInternalServerError)
		return
	}

	for {
		if err := r.Context().Err(); err != nil {
			break
		}

		var streams []string
		for _, forum := range fora {
			streams = append(streams, forum, since)
		}
		results, err := rdb.XRead(
			context.Background(),
			&redis.XReadArgs{
				Streams: streams,
				Count:   0,
				Block:   Keepalive,
			},
		).Result()
		if err == redis.Nil {
			// Keepalive timeout was hit with no data
			fmt.Fprintln(w, ": ping")
		} else if err != nil {
			log.Fatalf("XReadStreams(%v) => %v, %v", streams, results, err)
		}
		for _, res := range results {
			for _, rmsg := range res.Messages {
				var user string

				if val, ok := rmsg.Values["user"]; !ok {
					http.Error(w, fmt.Sprintf("user not defined on message %s", rmsg.ID), http.StatusInternalServerError)
					return
				} else {
					user = val.(string)
				}

				username, err := resolver.Resolve(event, user)
				if err != nil {
					username = fmt.Sprintf("??? %s", err.Error())
				}

				ucmsg := Message{
					User: username,
					Text: rmsg.Values["text"].(string),
				}
				jmsg, err := json.Marshal(ucmsg)
				if err != nil {
					http.Error(w, fmt.Sprintf("JSON Marshal: %s", err.Error()), http.StatusInternalServerError)
					return
				}
				fmt.Fprintf(w, "id: %s\n", rmsg.ID)
				fmt.Fprintf(w, "data: %s\n", string(jmsg))
				fmt.Fprintf(w, "\n")

				// next loop iteration, only ask for stuff that's happened since the last message
				since = rmsg.ID
			}
		}
		flusher.Flush()
	}
}

func main() {
	redisServer := flag.String("redis", "localhost:6379", "redis server")
	alfioAuth := flag.String("alfio", "", "Enable alfio authentication with given API base URL")
	hmacAuth := flag.String("hmac", "", "Enable HMAC authentication with given secret")
	noAuth := flag.Bool("noauth", false, "Enable lame (aka no) authentication")
	flag.Parse()

	rdb = redis.NewClient(&redis.Options{Addr: *redisServer})

	if *alfioAuth != "" {
		alfResolver := NewAlfioUserResolver(*alfioAuth)
		resolver = NewCacheResolver(alfResolver, rdb, 15*time.Minute)
	} else if *hmacAuth != "" {
		resolver = &HmacResolver{key: *hmacAuth}
	} else if *noAuth {
		resolver = NoAuthResolver{}
	} else {
		log.Fatal("No resolver specified")
		return
	}
	throttler = &Throttler{
		rdb:        rdb,
		expiration: 2 * time.Second,
	}

	http.HandleFunc("/say", sayHandler)
	http.HandleFunc("/read", readHandler)
	http.Handle("/", http.FileServer(http.Dir("static/")))

	bind := ":8080"
	log.Printf("Listening on %s", bind)
	log.Fatal(http.ListenAndServe(bind, nil))
}
