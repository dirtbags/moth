// XXX: Hack for chrome not supporting an iterator method on HTMLCollection
HTMLCollection.prototype[Symbol.iterator] = Array.prototype[Symbol.iterator];
NodeList.prototype[Symbol.iterator] = Array.prototype[Symbol.iterator];

function Terminal(target, bps) {
    bps = bps || 1200;

    var outq = [];
    var outTimer;

    function tx(nodes, bps, scroll) {
	var drawTimer;

	// Looks like EMCAScript 6 has a yield statement. That'll be nice.
	//
	// for (var node of nodes) {
	//   var text = "";
	//  for (var c of node._text) {
	//    text += c;
	//    node.textContent = text;
	//   }
	// }

	var nodeIndex = 0;
	var node = nodes[0];

	var textIndex = 0;
	var text = "";

	function draw() {
	    var src = node._text;
	    var c = src[textIndex];

	    text += c;
	    node.textContent = text;

	    textIndex += 1;
	    if (textIndex == src.length) {
		textIndex = 0;
		text = "";

		nodeIndex += 1;
		if (nodeIndex == nodes.length) {
		    clearInterval(drawTimer);
		    return;
		}
		node = nodes[nodeIndex];
	    }

	    if (scroll) {
		node.scrollIntoView();
	    }
	}

	// N81 uses 1 stop bit, and 1 parity bit.
	// That works out to exactly 10 bits per byte.
	msec = 10000 / bps;
	
	drawTimer = setInterval(draw, msec);
	draw();
    }


    function start() {
	if (! outTimer) {
	    outTimer = setInterval(drawElement, 150);
	}
    }


    function stop() {
	if (outTimer) {
	    clearInterval(outTimer);
	    outTimer = null;
	}
    }

    
    function drawElement() {
	var element = outq.shift();

	console.log(element);
	if (! element) {
	    stop();
	    return;
	}

	tx(element._terminalNodes, bps);
    }


    function prepare(element) {
	var nodes = [];

	walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
	while (walker.nextNode()) {
	    var node = walker.currentNode;
	    var text = node.textContent;

	    node.textContent = "";
	    nodes.push(node);
	}

	element._terminalNodes = nodes;
    }
	

    // The main entry point: works like appendChild
    this.append = function(element) {
	prepare(element);
	target.appendChild(element);
	outq.push(element);
	start();
    }


    // A cool effect where it despools children in parallel
    this.appendShallow = function(element) {
	for (var child of element.childNodes) {
	    prepare(child);
	    outq.push(child);
	}
	target.appendChild(element);
	start();
    }


    this.clear = function() {
	stop();
	outq = [];
	while (target.firstChild) {
	    target.removeChild(target.firstChild);
	}
    }


    this.par = function(txt) {
	var e = document.createElement("p");
	e.textContent = txt;
	this.append(e);
    }


    this.pre = function(txt) {
	var e = document.createElement("pre");
	e.textContent = txt;
	this.append(e);
    }
}

//
// Usage:
//
// var e = Terminal(document.getElementById("output"));
// e.output("This is a paragraph. It has sentences.");
// e.output("This is a second paragraph.");
//

