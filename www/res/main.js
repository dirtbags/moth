var main_terminal;

function Main(element) {
    var term = new Terminal(element);

    this.start = function() {
	term.clear();
	term.par("Main terminal.")
	term.par("This is the main terminal. In this terminal you will get your puzzle content and someplace to enter in possible answers. It's probably just going to pull the old URL, steal the body element, and submit it to a new Terminal method for slow-despooling of the content of text nodes.")
	term.par("One side-effect of the method I'm considering to slow-despool pre-written HTML is that inline images will load before the text. While not exactly what I had in mind for the feel of the thing, it may still be an interesting effect. I mean, if anything, text should render the *quickest*, so if we're going to turn everything on its head, why not make images pull in quicker than text.");
	term.par("Anyway.");
	term.par("Hopefully this demo illustrates how things are going to look.");
    }
}


function main_start() {
    main_terminal = new Main(document.getElementById("main"));
    setTimeout(main_terminal.start, 2500);
}

window.addEventListener("load", main_start);
