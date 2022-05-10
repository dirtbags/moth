package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Throttler provides a per-user timeout on posting
type Throttler struct {
	rdb        *redis.Client
	expiration time.Duration
}

// CanPost returns true if the given userID is okay to post
func (t *Throttler) CanPost(eventID string, userID string) bool {
	key := fmt.Sprintf("throttle:%s|%s", eventID, userID)
	setargs := t.rdb.SetArgs(
		context.TODO(),
		key,
		true,
		redis.SetArgs{
			Mode: "NX",
			TTL:  t.expiration,
		},
	)
	if err := setargs.Err(); err == redis.Nil {
		return false
	} else if err != nil {
		log.Print(err)
	}

	return true
}
