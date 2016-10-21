var messages_terminal;

function Messages(element) {
    var term = new Terminal(element);

    function start() {
	term.clear();
	term.par("Messages terminal");
	term.par("I've long wanted a way to communicate things to participants, like «yes, we're aware that JS 12 is broken, we are working on it», or «tanks category is now open!». This might have updates about people scoring points, or provide a chat service (although that has not historically been well-utilized).");
    }

    term.clear();
    term.par("Loading…");
    setTimeout(start, 500);
}


function messages_start() {
    messages_terminal = new Messages(document.getElementById("messages"));
}

window.addEventListener("load", messages_start);
