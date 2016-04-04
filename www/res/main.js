var main_proc;

function Main(element) {
    console.log(element);
    var term = new Terminal(element);

    this.start = function() {
	console.log(element);
	term.clear();
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main terminal");
	term.par("Main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main main terminal");
    }
}


function main_start() {
    main_proc = new Main(document.getElementById("main"));
    setTimeout(main_proc.start, 2500);
}

window.addEventListener("load", main_start);
