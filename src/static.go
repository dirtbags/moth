package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Status int

const (
	Success = iota
	Fail
	Error
)

// ShowJSend renders a JSend response to w
func ShowJSend(w http.ResponseWriter, status Status, short string, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // RFC2616 makes it pretty clear that 4xx codes are for the user-agent

	statusStr := ""
	switch status {
	case Success:
		statusStr = "success"
	case Fail:
		statusStr = "fail"
	default:
		statusStr = "error"
	}

	jshort, _ := json.Marshal(short)
	jdesc, _ := json.Marshal(description)
	fmt.Fprintf(
		w,
		`{"status":"%s","data":{"short":%s,"description":%s}}"`,
		statusStr, jshort, jdesc,
	)
}

// ShowHtml delevers an HTML response to w
func ShowHtml(w http.ResponseWriter, status Status, title string, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	statusStr := ""
	switch status {
	case Success:
		statusStr = "Success"
	case Fail:
		statusStr = "Fail"
	default:
		statusStr = "Error"
	}

	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<html><head>")
	fmt.Fprintf(w, "<title>%s</title>", title)
	fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"basic.css\">")
	fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\"></head>")
	fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/icon.svg\" type=\"image/svg+xml\">")
	fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/icon.png\" type=\"image/png\">")
	fmt.Fprintf(w, "<body><h1 class=\"%s\">%s</h1>", statusStr, title)
	fmt.Fprintf(w, "<section>%s</section>", body)
	fmt.Fprintf(w, "<nav>")
	fmt.Fprintf(w, "<ul>")
	fmt.Fprintf(w, "<li><a href=\"puzzles.html\">Puzzles</a></li>")
	fmt.Fprintf(w, "<li><a href=\"scoreboard.html\">Scoreboard</a></li>")
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</nav>")
	fmt.Fprintf(w, "</body></html>")
}

// staticStylesheet serves up a basic stylesheet.
// This is designed to be usable on small touchscreens (like mobile phones)
func staticStylesheet(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(
		w,
		`
/* http://paletton.com/#uid=63T0u0k7O9o3ouT6LjHih7ltq4c */
body {
  font-family: sans-serif;
  max-width: 40em;
	background: #282a33;
	color: #f6efdc;
}
a:any-link {
	color: #8b969a;
}
h1 {
	background: #5e576b;
	color: #9e98a8;
}
h1.Fail, h1.Error {
	background: #3a3119;
	color: #ffcc98;
}
h1.Fail:before {
	content: "Fail: ";
}
h1.Error:before {
	content: "Error: ";
}
p {
	margin: 1em 0em;
}
form, pre {
	margin: 1em;
}
input {
	padding: 0.6em;
	margin: 0.2em;
}
li {
	margin: 0.5em 0em;
}
nav {
  border: solid black 2px;
}
nav li {
	display: inline;
	margin: 2em;
}
		`,
	)
}

// staticIndex serves up a basic landing page
func staticIndex(w http.ResponseWriter) {
	ShowHtml(
		w, Success,
		"Welcome",
		`
<h2>Register your team</h2>

<form action="register" action="post">
  Team ID: <input name="id"> <br>
  Team name: <input name="name">
  <input type="submit" value="Register">
</form>

<p>
  If someone on your team has already registered,
  proceed to the
  <a href="puzzles.html">puzzles overview</a>.
</p>
		`,
	)
}

func staticScoreboard(w http.ResponseWriter) {
	ShowHtml(
		w, Success,
		"Scoreboard",
		"XXX: This would be the scoreboard",
	)
}

func staticPuzzles(w http.ResponseWriter) {
	ShowHtml(
		w, Success,
		"Open Puzzles",
		`
<div id="puzzles"></div>
<script>

function render(obj) {
	let element = document.getElementById("puzzles");
  let cats = [];
  for (let cat in obj) {
    cats.push(cat);
    console.log(cat);
  }
  cats.sort();

  for (let cat of cats) {
    let puzzles = obj[cat];
    
    let pdiv = document.createElement('div');
    pdiv.className = 'category';
    
    let h = document.createElement('h2');
    pdiv.appendChild(h);
    h.textContent = cat;
    
    let l = document.createElement('ul');
    pdiv.appendChild(l);
    
    for (var puzzle of puzzles) {
      var points = puzzle[0];
      var id = puzzle[1];
  
      var i = document.createElement('li');
      l.appendChild(i);
  
      if (points === 0) {
  	    i.textContent = "‡";
      } else {
      	var a = document.createElement('a');
      	i.appendChild(a);
      	a.textContent = points;
        a.href = cat + "/" + id + "/index.html";
	    }
  	}

		element.appendChild(pdiv);
  }
}

function init() {
	fetch("puzzles.json")
	.then(function(resp) {
		return resp.json();
	}).then(function(obj) {
		render(obj);
	}).catch(function(err) {
		console.log("Error", err);
	});
}

window.addEventListener("load", init);
</script>
		`,
	)
}

func tryServeFile(w http.ResponseWriter, req *http.Request, path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return false
	}

	http.ServeContent(w, req, path, d.ModTime(), f)
	return true
}

func ServeStatic(w http.ResponseWriter, req *http.Request, resourcesDir string) {
	path := req.URL.Path
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	if path == "/" {
		path = "/index.html"
	}

	fpath := filepath.Join(resourcesDir, path)
	if tryServeFile(w, req, fpath) {
		return
	}

	switch path {
	case "/basic.css":
		staticStylesheet(w)
	case "/index.html":
		staticIndex(w)
	case "/scoreboard.html":
		staticScoreboard(w)
	case "/puzzles.html":
		staticPuzzles(w)
	default:
		http.NotFound(w, req)
	}
}
