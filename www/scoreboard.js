function loadJSON(url, callback) {
	function loaded(e) {
		callback(e.target.response);
	}
	var xhr = new XMLHttpRequest()
	xhr.onload = loaded;
	xhr.open("GET", url, true);
	xhr.responseType = "json";
	xhr.send();
}

function scoreboard(element, continuous) {
	function update(state) {
		var teamnames = state["teams"];
		var pointslog = state["points"];
		var highscore = {};
		var teams = {};
		
		// Dole out points
		for (var i in pointslog) {
			var entry = pointslog[i];
			var timestamp = entry[0];
			var teamhash = entry[1];
			var category = entry[2];
			var points = entry[3];
			
			var team = teams[teamhash] || {__hash__: teamhash};
			
			// Add points to team's points for that category
			team[category] = (team[category] || 0) + points;
			
			// Record highest score in a category
			highscore[category] = Math.max(highscore[category] || 0, team[category]);
			
			teams[teamhash] = team;
		}
		
		// Sort by team score
		function teamScore(t) {
			var score = 0;

			for (var category in highscore) {
				score += (t[category] || 0) / highscore[category];
			}
			// XXX: This function really shouldn't have side effects.
			t.__score__ = score;
			return score;
		}
		function teamCompare(a, b) {
			return teamScore(a) - teamScore(b);
		}
		
		var winners = [];
		for (var i in teams) {
			winners.push(teams[i]);
		}
		if (winners.length == 0) {
			// No teams!
			return;
		}
		winners.sort(teamCompare);
		winners.reverse();
		
		// Clear out the element we're about to populate
		while (element.lastChild) {
			element.removeChild(element.lastChild);
		}
		
		// Populate!
		var topActualScore = winners[0].__score__;

		// (100 / ncats) * (ncats / topActualScore);
		var maxWidth = 100 / topActualScore;
		for (var i in winners) {
			var team = winners[i];
			var row = document.createElement("div");
			var ncat = 0;
			for (var category in highscore) {
				var catHigh = highscore[category];
				var catTeam = team[category] || 0;
				var catPct = catTeam / catHigh;
				var width = maxWidth * catPct;
				
				var bar = document.createElement("span");
				bar.classList.add("cat" + ncat);
				bar.style.width = width + "%";
				bar.textContent = category + ": " + catTeam;
				bar.title = bar.textContent;
				
				row.appendChild(bar);
				ncat += 1;
			}
			
			var te = document.createElement("span");
			te.classList.add("teamname");
			te.textContent = teamnames[team.__hash__];
			row.appendChild(te);
			
			element.appendChild(row);
		}
	}
	
	function once() {
		loadJSON("points.json", update);
	}
	if (continuous) {
		setInterval(once, 60000);
	}
	once();
}
