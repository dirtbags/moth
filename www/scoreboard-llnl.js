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

function toObject(arr) {
  var rv = {};
  for (var i = 0; i < arr.length; ++i)
    if (arr[i] !== undefined) rv[i] = arr[i];
  return rv;
}

var updateInterval;

function scoreboard(element, continuous, mode, interval) {
	if(updateInterval)
	{
		clearInterval(updateInterval);
	}
	function update(state) {
		//console.log("Updating");
		var teamnames = state["teams"];
		var pointslog = state["points"];
		var highscore = {};
		var teams = {};
		
		function pointsCompare(a, b) {
			return a[0] - b[0];
		}
		pointslog.sort(pointsCompare);
		var minTime = pointslog[0][0];
		var maxTime = pointslog[pointslog.length - 1][0];

		var allQuestions = {};
		
		for (var i in pointslog)
		{
			var entry = pointslog[i];
			var timestamp = entry[0];
			var teamhash = entry[1];
			var category = entry[2];
			var points = entry[3];
			
			var catPoints = {};
			if(category in allQuestions)
			{
				catPoints = allQuestions[category];
			}
			else
			{
				catPoints["total"] = 0;
			}
			
			if(!(points in catPoints))
			{
				catPoints[points] = 1;
				catPoints["total"] = catPoints["total"] + points;
			}
			else
			{
				catPoints[points] = catPoints[points] + 1;
			}
			
			allQuestions[category] = catPoints;
		}

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
		function pointScore(points, category)
		{
			return points / highscore[category]
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
		

		if(mode == "time")
		{
			var colorScale = d3.schemeCategory20;
			
			var teamLines = {};
			var reverseTeam = {};
			for(var i in pointslog)
			{
				var entry = pointslog[i];
				var timestamp = entry[0];
				var teamhash = entry[1];
				var category = entry[2];
				var points = entry[3];
				var teamname = teamnames[teamhash];
				reverseTeam[teamname] = teamhash;
				points = pointScore(points, category);

				if(!(teamname in teamLines))
				{
					var teamHistory = [[timestamp, points, category, entry[3], [minTime, 0, category, 0]]];
					teamLines[teamname] = teamHistory;
				}
				else
				{
					var teamHistory = teamLines[teamname];
					teamHistory.push([timestamp, points + teamHistory[teamHistory.length - 1][1], category, entry[3], teamHistory[teamHistory.length - 1]]);
				}
			}

			//console.log(teamLines);
			
			var graph = document.createElement("svg");
			graph.id = "graph";
			graph.style.width="100%";
			graph.style.height="40em";
			graph.style.backgroundColor = "white";
			graph.style.display = "table";
			var holdingDiv = document.createElement("div");
			holdingDiv.align="center";
			holdingDiv.id="holding";
			element.appendChild(holdingDiv);
			holdingDiv.appendChild(graph);
			
			var margins = 40;
			var marginsX = 120;

			var width = graph.offsetWidth;
			var height = graph.offsetHeight;

			//var xScale = d3.scaleLinear().range([minTime, maxTime]);
			//var yScale = d3.scaleLinear().range([0, topActualScore]);
			var originTime = (maxTime - minTime) / 60;
			var xScale = d3.scaleLinear().range([marginsX, width - margins]);
			xScale.domain([0, originTime]);
			var yScale = d3.scaleLinear().range([height - margins, margins]);
			yScale.domain([0, topActualScore]);

			graph = d3.select("#graph");
			graph.remove();
			graph = d3.select("#holding").append("svg")
				.attr("width", width)
				.attr("height", height);
				//.attr("style", "background: white");
			
			
			//graph.append("g")
			//	.attr("transform", "translate(" + margins + ", 0)")
			//	.call(d3.axisLeft(yScale))
			//	.style("stroke", "white");;
				
			var maxNumEntry = 10;
			//var curEntry = 0;
			var winningTeams = [];
			for(entry in winners)
			{
				var curEntry = entry;
				if(curEntry >= maxNumEntry)
				{
					break;
				}
				entry = teamnames[winners[entry].__hash__];
				winningTeams.push(entry);
				//console.log(curEntry);
				//console.log(entry);
				
				//var isTop = false;
				//for(var x=0; x < maxNumEntry; x++)
				//{
				//	var teamhash = reverseTeam[entry];
				//	if(winners[x].__hash__ == teamhash)
				//	{
				//		curEntry = x;
				//		isTop = true;
				//		break;
				//	}
				//}
				//if(!isTop)
				//{
				//	continue;
				//}

				var curTeam = teamLines[entry];
				var lastEntry = curTeam[curTeam.length - 1];
				//curTeam.append()
				curTeam.push([maxTime, lastEntry[1], lastEntry[2], lastEntry[3], lastEntry]);
				var curLayer = graph.append("g");
				curLayer.selectAll("line")
					.data(curTeam)
					.enter()
					.append("line")
					.style("stroke", colorScale[curEntry * 2])
					.attr("stroke-width", 4)
					.attr("class", "team_" + entry)
					.style("z-index", maxNumEntry - curEntry)
					.attr("x1",
						function(d)
						{
							return xScale((d[4][0] - minTime) / 60);
						})
					.attr("x2",
						function(d)
						{
							return xScale((d[0] - minTime) / 60);
						})
					.attr("y1",
						function(d)
						{
							return yScale(d[4][1]);
						})
					.attr("y2",
						function(d)
						{
							return yScale(d[1]);
						})
					.on("mouseover", handleMouseover)
					.on("mouseout", handleMouseout);
				
				curLayer.selectAll("circle")
					.data(curTeam)
					.enter()
					.append("circle")
					.style("fill", colorScale[curEntry * 2])
					.style("z-index", maxNumEntry - curEntry)
					.attr("class", "team_" + entry)
					.attr("r", 5)
					.attr("cx",
						function(d)
						{
							return xScale((d[0] - minTime) / 60);
						})
					.attr("cy",
						function(d)
						{
							return yScale(d[1]);
						})
					.on("mouseover", handleMouseoverCircle)
					.on("mouseout", handleMouseoutCircle);
				
				curEntry++;
			}

			var axisG = graph.append("g");
			axisG
				.attr("transform", "translate(0," + (height - margins) + ")")
				.call(d3.axisBottom(xScale));
				//.style("stroke", "white");
			axisG.selectAll("path").style("stroke", "white");
			axisG.selectAll("line").style("stroke", "white");
			axisG.selectAll("text").style("fill", "white");
			
			graph.append("text")
				.attr("text-anchor", "middle")
				.attr("transform", "translate(" + (width / 2) + ", " + (height - margins / 8) + ")")
				.style("fill", "white")
				.text("Time (minutes)");

			var legend = graph.append("g");
			var legendRowHeight = (height) / 10;
			legend.selectAll("rect")
				.data(winningTeams)
				.enter()
				.append("rect")
				.attr("class", function(d){ return "team_" + d; })
				.attr("fill", function(d, i){ return colorScale[i * 2 + 1]; })
				.style("z-index", function(d, i){ return i; })
				.attr("x", 0)
				.attr("y", function(d, i){ return legendRowHeight * i; })
				.attr("height", legendRowHeight)
				.attr("width", marginsX)
				.on("mouseover", handleMouseoverLegend)
				.on("mouseout", handleMouseoutLegend);

			legend.selectAll("text")
				.data(winningTeams)
				.enter()
				.append("text")
				//.attr("class", function(d){ return "team_" + d; })
				.attr("fill", "black")
				.style("z-index", function(d, i){ return i; })
				.attr("dx", 0)
				.attr("dy", function(d, i){ return legendRowHeight * (i + .5); })
				.text(function(d, i){ return i + ": " + d; })
				.attr("dominant-baseline", "central")
				.style("pointer-events", "none");

			//legend.append("g").selectAll("text")
			//	.data(winningTeams)
			//	.enter()
			//	.append("text")
			//	.attr("class", function(d){ return "team_" + d; })
			//	.attr("fill", function(d, i){ return colorScale[i]; })
			//	.style("z-index", function(d, i){ return i; })
			//	.attr("dx", margins)
			//	.attr("dy", function(d, i){ return margins + legendRowHeight * (i); })
			//	.text(function(d){ return d; });
				//.attr("dominant-baseline", "central");
				//.style("pointer-events", "none");
				

			function handleMouseover(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", "white")
						.style("fill", "white");
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
				}

			function handleMouseout(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					var zIndex = d3.select(this).style("z-index");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", colorScale[(maxNumEntry - zIndex) * 2])
						.style("fill", colorScale[(maxNumEntry - zIndex) * 2]);
					legend.selectAll("." + curClass)
						.style("stroke", colorScale[(maxNumEntry - zIndex) * 2 + 1])
						.style("fill", colorScale[(maxNumEntry - zIndex) * 2 + 1]);
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
				}
			
			var tooltipPadding = 10;
			function handleMouseoverCircle(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", "white")
						.style("fill", "white");
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
					
					graph.append("g").append("text")
						.attr("class", "tooltip")
						.attr("text-anchor", "middle")
						.style("fill", "red")
						.style("stroke-width", -4)
						.style("stroke", "black")
						.style("font-weight", "bolder")
						.style("font-size", "large")
						.attr("dx",
							function()
							{
								return xScale((d[0] - minTime) / 60);
							})
						.attr("dy",
							function()
							{
								return yScale(d[1]) - tooltipPadding;
							})
						.text(function(){ return d[2] + " " + d[3]; })
						.style("pointer-events", "none");
					
				}

			function handleMouseoutCircle(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					var zIndex = d3.select(this).style("z-index");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", colorScale[(maxNumEntry - zIndex) * 2])
						.style("fill", colorScale[(maxNumEntry - zIndex) * 2]);
					legend.selectAll("." + curClass)
						.style("stroke", colorScale[(maxNumEntry - zIndex) * 2 + 1])
						.style("fill", colorScale[(maxNumEntry - zIndex) * 2 + 1]);
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
				}

			function handleMouseoverLegend(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", "white")
						.style("fill", "white");
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
				}

			function handleMouseoutLegend(d, i)
				{
					d3.select("body").selectAll(".tooltip").remove();
					var curClass = d3.select(this).attr("class");
					var zIndex = d3.select(this).style("z-index");
					d3.select("body").selectAll("." + curClass)
						.style("stroke", colorScale[zIndex * 2])
						.style("fill", colorScale[zIndex * 2]);
					legend.selectAll("." + curClass)
						.style("stroke", colorScale[(zIndex) * 2 + 1])
						.style("fill", colorScale[(zIndex) * 2 + 1]);
					d3.select("body").selectAll("text")
						.style("stroke-width", 0);
				}


		}
		else if(mode == "original")
		{
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
		if(mode == "total")
		{
			var colorScale = d3.schemeCategory20;
			
			var numCats = 0;
			for(entry in allQuestions)
			{
				numCats++;
			}
			var maxWidth = Math.floor(100 / (0.0 + numCats));
			//console.log(maxWidth);

			for (var i in winners) {
				var team = winners[i];
				var row = document.createElement("div");
				var ncat = 0;
				for (var category in allQuestions) {
					var catHigh = highscore[category];
					var catTeam = team[category] || 0;
					var catPct = (0.0 + catTeam) / (0.0 + catHigh);
					var width = maxWidth * catPct;
					var bar = document.createElement("span");
					
					var numLeft = catHigh - catTeam;
					
					//bar.classList.add("cat" + ncat);
					bar.style.backgroundColor = colorScale[ncat % 20];
					bar.style.color = "white";
					bar.style.width = width + "%";
					bar.textContent = category + ": " + catTeam;
					bar.title = bar.textContent;
					
					row.appendChild(bar);
					
					ncat++;
					
					width = maxWidth * (1 - catPct);
					if(width > 0)
					{
						var noBar = document.createElement("span");
						//noBar.classList.add("cat" + ncat);
						noBar.style.backgroundColor = colorScale[ncat % 20];
						noBar.style.width = width + "%";
						noBar.textContent = numLeft;
						noBar.title = bar.textContent;
				
						row.appendChild(noBar);
					}
					ncat += 1;
				}
			
				var te = document.createElement("span");
				te.classList.add("teamname");
				te.textContent = teamnames[team.__hash__];
				row.appendChild(te);
			
				element.appendChild(row);
			}
		}
	}
	
	function once() {
		loadJSON("points.json", update);
	}
	if (continuous) {
		updateInterval = setInterval(once, interval);
	}
	once();
}

