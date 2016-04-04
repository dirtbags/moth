var puzzles_proc;

function Puzzles(element) {
    var term = new Terminal(element);

    this.start = function() {
	term.clear();
	term.par("Puzzles terminal");
	term.par("This is going to show you the list of open puzzles. It should refresh itself periodically, since not refreshing was a source of major confusion in the last setup, at least for kids, who seem not to realize what the reload button in the browser does.")
    }
}


function puzzles_start() {
    puzzles_proc = new Puzzles(document.getElementById("puzzles"));
    setTimeout(puzzles_proc.start, 3000);
}

window.addEventListener("load", puzzles_start);
