var overview_proc;

function Overview(element) {
    var term = new Terminal(element);

    this.start = function() {
	term.clear();
	term.par("Overview terminal");
	term.par("Here you will find something resembling a scoreboard, maybe your team name.")
    }
}


function overview_start() {
    overview_proc = new Overview(document.getElementById("overview"));
    setTimeout(overview_proc.start, 4000);
}

window.addEventListener("load", overview_start);
