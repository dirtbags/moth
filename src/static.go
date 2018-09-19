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
.Fail, .Error {
	background: #3a3119;
	color: #ffcc98;
}
.Fail:before {
	content: "Fail: ";
}
.Error:before {
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
nav {
  border: solid black 2px;
}
nav ul, .category ul {
  padding: 1em;
}
nav li, .category li {
	display: inline;
	margin: 1em;
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
		`
    <section>
      <div id="scoreboard"></div>
    </section>
    <script>
function loadJSON(url, callback) {
	function loaded(e) {
		callback(e.target.response);
	}
	var xhr = new XMLHttpRequest()
	xhr.onload = loaded;
	xhr.open("GET", url, true);
	xhr.responseType = "json";
	xhr.send();
}

function scoreboard(element, continuous) {
	function update(state) {
		var teamnames = state["teams"];
		var pointslog = state["points"];
		var pointshistory = JSON.parse(localStorage.getItem("pointshistory")) || [];
		if (pointshistory.length >= 20){
			pointshistory.shift();
		}
		pointshistory.push(pointslog);
		localStorage.setItem("pointshistory", JSON.stringify(pointshistory));
		var highscore = {};
		var teams = {};

		// Dole out points
		for (var i in pointslog) {
			var entry = pointslog[i];
			var timestamp = entry[0];
			var teamhash = entry[1];
			var category = entry[2];
			var points = entry[3];

			var team = teams[teamhash] || {__hash__: teamhash};

			// Add points to team's points for that category
			team[category] = (team[category] || 0) + points;

			// Record highest score in a category
			highscore[category] = Math.max(highscore[category] || 0, team[category]);

			teams[teamhash] = team;
		}

		// Sort by team score
		function teamScore(t) {
			var score = 0;

			for (var category in highscore) {
				score += (t[category] || 0) / highscore[category];
			}
			// XXX: This function really shouldn't have side effects.
			t.__score__ = score;
			return score;
		}
		function teamCompare(a, b) {
			return teamScore(a) - teamScore(b);
		}

		var winners = [];
		for (var i in teams) {
			winners.push(teams[i]);
		}
		if (winners.length == 0) {
			// No teams!
			return;
		}
		winners.sort(teamCompare);
		winners.reverse();

		// Clear out the element we're about to populate
		while (element.lastChild) {
			element.removeChild(element.lastChild);
		}

		// Populate!
		var topActualScore = winners[0].__score__;

		// (100 / ncats) * (ncats / topActualScore);
		var maxWidth = 100 / topActualScore;
		for (var i in winners) {
			var team = winners[i];
			var row = document.createElement("div");
			var ncat = 0;
			for (var category in highscore) {
				var catHigh = highscore[category];
				var catTeam = team[category] || 0;
				var catPct = catTeam / catHigh;
				var width = maxWidth * catPct;

				var bar = document.createElement("span");
				bar.classList.add("cat" + ncat);
				bar.style.width = width + "%";
				bar.textContent = category + ": " + catTeam;
				bar.title = bar.textContent;

				row.appendChild(bar);
				ncat += 1;
			}

			var te = document.createElement("span");
			te.classList.add("teamname");
			te.textContent = teamnames[team.__hash__];
			row.appendChild(te);

			element.appendChild(row);
		}
	}

	function once() {
		loadJSON("points.json", update);
	}
	if (continuous) {
		setInterval(once, 60000);
	}
	once();
}

function init() {
	var sb = document.getElementById("scoreboard");
	scoreboard(sb, true);
}

document.addEventListener("DOMContentLoaded", init);
    </script>
		`,
	)
}

func staticPuzzles(w http.ResponseWriter) {
	ShowHtml(
		w, Success,
		"Open Puzzles",
		`
<section>
	<div id="puzzles"></div>
</section>
<script>
function render(obj) {
  puzzlesElement = document.createElement('div');
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

		puzzlesElement.appendChild(pdiv);
		document.getElementById("puzzles").appendChild(puzzlesElement);
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

document.addEventListener("DOMContentLoaded", init);
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
