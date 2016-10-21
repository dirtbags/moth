var puzzlesTerminal;
var puzzlesJsonUrl = "puzzles.json";

function loadPuzzle(cat, id, points) {
    console.log("Requested " + cat + "/" + id + "(" + points + ")");
}

function puzzlesRefresh(term, obj) {
    term.clear();

    for (var cat in obj) {
	var puzzles = obj[cat];

	var pdiv = createElement('div');
	pdiv.className = 'category';

	var h = createElement('h2');
	pdiv.appendChild(h);
	h.textContent = cat;

	var l = createElement('ul');
	pdiv.appendChild(l);

	for (var puzzle of puzzles) {
	    var points = puzzle[0];
	    var id = puzzle[1];

	    var i = createElement('li');
	    l.appendChild(i);

	    if (points == 0) {
		i.textContent = "‡";
	    } else {
		var a = createElement('span');
		i.appendChild(a);
		a.className = "link";
		a.textContent = points;
		a.addEventListener("click", loadPuzzle.bind(undefined, cat, id, points));
	    }
	}

	term.appendShallow(pdiv);
    }
}

function puzzles_start() {
    var element = document.getElementById("puzzles");
    var puzzlesTerminal = new Terminal(element);
    var refreshInterval = 40 * 1000;

    var refreshCallback = puzzlesRefresh.bind(undefined, puzzlesTerminal);
    var refreshFunction = loadJSON.bind(undefined, puzzlesJsonUrl, refreshCallback);

    puzzlesTerminal.clear();
    puzzlesTerminal.par("Loading...");
    refreshFunction();
    setInterval(refreshFunction, refreshInterval);
}

window.addEventListener("load", puzzles_start);
