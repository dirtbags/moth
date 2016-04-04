function tx(element, text, bps) {
    var drawTimer;

    var sp = false;
    function draw() {
	c = text[0];
	if ((c == " ") || (c == "\n")) {
	    sp = true;
	    c = " ";
	} else if (sp) {
	    c = " " + c;
	    sp = false;
	}
	element.textContent += c;
	text = text.substr(1);
	if (text == "") {
	    clearInterval(drawTimer);
	    return;
	}
    }

    // N81 uses 1 stop bit, and 1 parity bit.
    // That works out to exactly 10 bits per byte.
    msec = 10000 / bps;
	
    drawTimer = setInterval(draw, msec);
    draw();
}

function Terminal(target, bps) {
    bps = bps || 1200;

    var outq = [];
    var outTimer;

    function drawElement() {
	var next = outq.shift();
	var out = document.createElement(next[0]);

	target.appendChild(out);
	tx(out, next[1], bps);

	if (outq.length == 0) {
	    clearInterval(outTimer);
	}
    }

    this.enqueue = function(tag, txt) {
	outq.push([tag, txt]);
	if (! outTimer) {
	    outTimer = setInterval(drawElement, 150);
	}
    }

    this.par = function(txt) {
	this.enqueue("p", txt);
    }

    this.pre = function(txt) {
	this.enqueue("pre", txt);
    }
}

//
// Usage:
//
// var e = Terminal(document.getElementById("output"));
// e.output("This is a paragraph. It has sentences.");
// e.output("This is a second paragraph.");
//

