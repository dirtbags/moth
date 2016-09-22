var puzzles_terminal;

var puzzles_url = "hack/puzzles.html";

function Puzzles(element) {
    var term = new Terminal(element);
    var refreshInterval;

    function loaded() {
	var doc = this.response;
	var puzzles = doc.getElementById("puzzles");
	var h1 = document.createElement("h1");

	h1.textContent = "Puzzles";

	term.clear();
	term.append(h1);
	term.appendShallow(puzzles);
    }

    function refresh() {
	var myRequest = new XMLHttpRequest();
	myRequest.responseType = "document";
	myRequest.addEventListener("load", loaded);
	myRequest.open("GET", puzzles_url);
	myRequest.send();
    }

    function start() {
	refreshInterval = setInterval(refresh, 20 * 1000);
	refresh();
    }

    term.clear();
    term.par("Loading…");

    setTimeout(start, 3000);
}


function puzzles_start() {
    puzzles_terminal = new Puzzles(document.getElementById("puzzles"));
}

window.addEventListener("load", puzzles_start);
