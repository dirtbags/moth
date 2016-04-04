var messages_proc;

function Messages(element) {
    var term = new Terminal(element);

    this.start = function() {
	term.clear();
	term.par("Messages terminal");
    }
}


function messages_start() {
    messages_proc = new Messages(document.getElementById("messages"));
    setTimeout(messages_proc.start, 500);
}

window.addEventListener("load", messages_start);
