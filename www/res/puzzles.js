var puzzlesTerminal;
var puzzlesJsonUrl = "puzzles.json";

function loadPuzzle(cat, id, points) {
    console.log("Requested " + cat + "/" + id + "(" + points + ")");
}

function puzzlesRefresh(term, obj) {
  term.clear();
  
  let cats = [];
  for (let cat in obj) {
    cats.push(cat);
  }
  cats.sort();

  for (let cat of cats) {
    let puzzles = obj[cat];
    
    let pdiv = createElement('div');
    pdiv.className = 'category';
    
    let h = createElement('h2');
    pdiv.appendChild(h);
    h.textContent = cat;
    
    let l = createElement('ul');
    pdiv.appendChild(l);
    
    for (var puzzle of puzzles) {
      var points = puzzle[0];
      var id = puzzle[1];
  
      var i = createElement('li');
      l.appendChild(i);
  
      if (points === 0) {
  	    i.textContent = "‡";
      } else {
      	var a = createElement('a');
      	i.appendChild(a);
      	a.className = "link";
      	a.textContent = points;
        a.href = cat + "/" + id + "/index.html";
    	// a.addEventListener("click", loadPuzzle.bind(undefined, cat, id, points));
	    }
  	}

	term.appendShallow(pdiv);
  }
}

function puzzles_start() {
  var element = document.getElementById("puzzles");
  var puzzlesTerminal = new Terminal(element);
  var refreshInterval = 40 * 1000;

  var refreshCallback = puzzlesRefresh.bind(undefined, puzzlesTerminal);
  var refreshFunction = loadJSON.bind(undefined, puzzlesJsonUrl, refreshCallback);

  puzzlesTerminal.clear();
  puzzlesTerminal.par("Loading...");
  refreshFunction();
  setInterval(refreshFunction, refreshInterval);
}

window.addEventListener("load", puzzles_start);
