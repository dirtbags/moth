var baud = 1200;

function tx(element, text, baud_) {
	var drawTimer;
	baud_ = baud_ || baud;
	
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
	msec = 10000 / baud_;
	
	drawTimer = setInterval(draw, msec);
	draw();
}


var outq = [];
var outTimer;

function drawPar() {
	oute = document.getElementById("output");
	outp = document.createElement("p");

	oute.appendChild(outp);
	tx(outp, outq.shift());
	if (outq.length == 0) {
		clearInterval(outTimer);
	}
}

function output(par) {
	outq = outq.concat(par);
	if (! outTimer) {
		outTimer = setInterval(drawPar, 150);
	}
}

