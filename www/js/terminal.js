// A class to turn an element into a cybersteampunk terminal.
// Runs at 1200 baud by default, but unlike an actual modem,
// will despool in parallel. This looks pretty cool.

// XXX: Hack for chrome not supporting an iterator method on HTMLCollection
HTMLCollection.prototype[Symbol.iterator] = Array.prototype[Symbol.iterator];
NodeList.prototype[Symbol.iterator] = Array.prototype[Symbol.iterator];

function Terminal(target, bps) {
    bps = bps || 9600;

    var outq = [];
    var outTimer;

    // Heavy lifting happens here.
    // At first I had it auto-scrolling to the bottom, like xterm (1987).
    // But that was actually kind of annoying, since this is meant to be read.
    // So now it leaves the scrollbar in place, and the user has to scroll.
    // This is how the Plan 9 terminal (1991) works.
    function tx(pairs, bps, scroll) {
      var drawTimer;
      
      // Looks like EMCAScript 6 has a yield statement.
      // That would make this mess a lot easier to understand.
      
      var pairIndex = 0;
      var pair = pairs[0];
      
      var textIndex = 0;
      var text = "";
      
      function draw() {
          var node = pair[0];
          var src = pair[1];
          var c = src[textIndex];
      
          text += c;
          node.textContent = text;
      
          textIndex += 1;
          if (textIndex == src.length) {
            textIndex = 0;
            text = "";
            
            pairIndex += 1;
            if (pairIndex == pairs.length) {
                clearInterval(drawTimer);
                return;
            }
          pair = pairs[pairIndex];
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
	    outTimer = setInterval(drawElement, 25);
	}
    }


    function stop() {
	if (outTimer) {
	    clearInterval(outTimer);
	    outTimer = null;
	}
    }

    
    function drawElement() {
	var pairs = outq.shift();

	if (! pairs) {
	    stop();
	    return;
	}

	tx(pairs, bps);
    }


    function prepare(element) {
	var pairs = [];

	walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
	while (walker.nextNode()) {
	    var node = walker.currentNode;
	    var text = node.textContent;

	    node.textContent = "";
	    pairs.push([node, text]);
	}

	return pairs;
    }
	

    // The main entry point: works like appendChild
    this.append = function(element) {
	pairs = prepare(element);
	target.appendChild(element);
	outq.push(pairs);
	start();
    }


    // A cool effect where it despools children in parallel
    this.appendShallow = function(element) {
	for (var child of element.childNodes) {
	    pairs = prepare(child);
	    outq.push(pairs);
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

