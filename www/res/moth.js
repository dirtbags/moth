// Moth dashboard
// requires: terminal.js

function start() {
    var t = new Terminal(document.getElementById("output"));

    t.par("This is a paragraph, bitches!");
    t.pre("This is pre");
    t.par("Another par");
}

window.addEventListener("load", start);
