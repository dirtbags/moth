var puzzles_proc;

function Puzzles(element) {
    var term = new Terminal(element);

    this.start = function() {
	term.clear();
	term.par("Puzzles terminal");
    }
}


function puzzles_start() {
    puzzles_proc = new Puzzles(document.getElementById("puzzles"));
    setTimeout(puzzles_proc.start, 3000);
}

window.addEventListener("load", puzzles_start);
