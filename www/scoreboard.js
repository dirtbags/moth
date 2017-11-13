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

function scoreboardHistoryPush(pointslog) {
    let pointsHistory = JSON.parse(localStorage.getItem("pointsHistory")) || [];
    if (pointsHistory.length >= 20) {
	pointsHistory.shift();
    }
    pointsHistory.push(pointslog);
    localStorage.setItem("pointsHistory", JSON.stringify(pointsHistory));
}

function scoreboard(element, continuous) {
    function update(state) {
	let teamNames = state["teams"];
	let pointsLog = state["points"];

        // Establish scores, calculate category maximums
        let categories = {};
        let maxPointsByCategory = {};
        let totalPointsByTeamByCategory = {};
        for (let entry of pointsLog) {
            let entryTimeStamp = entry[0];
            let entryTeamHash = entry[1];
            let entryCategory = entry[2];
            let entryPoints = entry[3];

            // Populate list of all categories
            categories[entryCategory] = entryCategory;
            
            // Add points to team's points for that category
            let points = totalPointsByTeamByCategory[entryTeamHash] || {};
            let categoryPoints = points[entryCategory] || 0;
            categoryPoints += entryPoints;
            points[entryCategory] = categoryPoints;
            totalPointsByTeamByCategory[entryTeamHash] = points;

            // Calculate maximum points scored in each category
            let m = maxPointsByCategory[entryCategory] || 0;
            maxPointsByCategory[entryCategory] = Math.max(m, categoryPoints);
        }

        // Calculate overall scores
        let overallScore = {};
        let orderedOverallScores = [];
	for (let teamHash in teamNames) {
	    var score = 0;
            for (let cat in categories) {
		var catPoints = totalPointsByTeamByCategory[teamHash][cat] || 0;
		if (catPoints > 0) {
                    score += catPoints / maxPointsByCategory[cat];
		}
            }
            overallScore[teamHash] = score;
            orderedOverallScores.push([score, teamHash]);
        }
        orderedOverallScores.sort();
	orderedOverallScores.reverse();

	// Clear out the element we're about to populate
	while (element.lastChild) {
	    element.removeChild(element.lastChild);
	}

	// Set up scoreboard structure
	let spansByTeamByCategory = {};
	for (let pair of orderedOverallScores) {
	    let teamHash = pair[1];
	    let teamName = teamNames[teamHash];
	    let teamRow = document.createElement("div");
	    let ncat = 0;
	    spansByTeamByCategory[teamHash] = {};
	    for (let cat in categories) {
		let catSpan = document.createElement("span");
		catSpan.classList.add("cat" + ncat);
		catSpan.style.width = "0%";
		catSpan.textContent = cat + ": 0";

		spansByTeamByCategory[teamHash][cat] = catSpan;

		teamRow.appendChild(catSpan);
		ncat += 1;
	    }

	    var te = document.createElement("span");
	    te.classList.add("teamname");
	    te.textContent = teamName;
	    teamRow.appendChild(te);

	    element.appendChild(teamRow);
	}

	// How many categories are there?
	var numCategories = 0;
	for (var cat in categories) {
	    numCategories += 1;
	}

	// Replay points log, displaying scoreboard at each step
	let replayTimer = null;
	let replayIndex = 0;
	function replayStep(event) {
	    if (replayIndex > pointsLog.length) {
		clearInterval(replayTimer);
		return;
	    }

	    // Replay log up until replayIndex
	    let totalPointsByTeamByCategory = {};
	    for (let index = 0; index < replayIndex; index += 1) {
		let entry = pointsLog[index];
		let entryTimeStamp = entry[0];
		let entryTeamHash = entry[1];
		let entryCategory = entry[2];
		let entryPoints = entry[3];

		// Add points to team's points for that category
		let points = totalPointsByTeamByCategory[entryTeamHash] || {};
		let categoryPoints = points[entryCategory] || 0;
		categoryPoints += entryPoints;
		points[entryCategory] = categoryPoints;
		totalPointsByTeamByCategory[entryTeamHash] = points;
	    }

	    // Figure out everybody's score
	    for (let teamHash in teamNames) {
		for (let cat in categories) {
		    let totalPointsByCategory = totalPointsByTeamByCategory[teamHash] || {};
		    let points = totalPointsByCategory[cat] || 0;
		    if (points > 0) {
			let score = points / maxPointsByCategory[cat];
			let span = spansByTeamByCategory[teamHash][cat];
			let width = (100.0 / numCategories) * score;

			span.style.width = width + "%";
			span.textContent = cat + ": " + points;
			span.title = span.textContent;
		    }
		}
	    }

	    replayIndex += 1;
	}
	replayStep();
	replayTimer = setInterval(replayStep, 20);
    }

    function once() {
	loadJSON("points.json", update);
    }
    if (continuous) {
	setInterval(once, 60000);
    }
    once();
}

