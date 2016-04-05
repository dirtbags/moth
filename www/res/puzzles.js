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
	term.append(puzzles);
    }

    function refresh() {
	var myRequest = new XMLHttpRequest();
	myRequest.responseType = "document";
	myRequest.addEventListener("load", loaded);
	myRequest.open("GET", puzzles_url);
	myRequest.send();
    }

    function start() {
	term.clear();
	term.par("Loading...");

	term.par("This is going to show you the list of open puzzles. It should refresh itself periodically, since not refreshing was a source of major confusion in the last setup, at least for kids, who seem not to realize what the reload button in the browser does.")

	refreshInterval = setInterval(refresh, 20 * 1000);
	refresh();
    }

    setTimeout(start, 3000);
}


function puzzles_start() {
    puzzles_terminal = new Puzzles(document.getElementById("puzzles"));
}

window.addEventListener("load", puzzles_start);
