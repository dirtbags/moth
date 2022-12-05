// jshint asi:true

var MOTH_RANKING_STANDARD=0;
var MOTH_RANKING_CATEGORY=1;
var MOTH_RANKING_TRACK=2;

// Comparison functions
var MOTH_COMP_TEAMOVERALL=function (a, b) {
  return a.overallScore - b.overallScore;
}

var MOTH_COMP_POINTSLOGTIME=function(a, b) {
  return a[0] - b[0];
}

var MOTH_COMP_SCORE=function(a, b) {
  return a - b;
}

var teamNames={};
var pointsLog={};

function scoreboardInit() {
  // Visual flare
  var cortex={
    ranking: MOTH_RANKING_STANDARD,
    sorting: MOTH_COMP_TEAMOVERALL,
    screenSaverBg: "svg",
  };
  
  var chartColors = [
    "rgb(255, 99, 132)",
    "rgb(255, 159, 64)",
    "rgb(255, 205, 86)",
    "rgb(75, 192, 192)",
    "rgb(54, 162, 235)",
    "rgb(153, 102, 255)",
    "rgb(201, 203, 207)"
  ]

  // Placeholder for Track mappings
  trackMap = {
    "Operational-Technology" : "ot",
    "Safe_Malware" : "malware",
    "linux_memory_intro" : "forensics",
    "js" : "netarch",
    "sequence" : "netarch",
    "networking" : "entry-point",
    "codebreaking" : "incident-coordination",
    "nocode" : "entry-point",
  }

  function update(state) {
    window.state = state
    
    for (let rotate of document.querySelectorAll(".rotate")) {
      rotate.appendChild(rotate.firstElementChild)
    }
    
    teamNames = state.TeamNames
    pointsLog = state.PointsLog

    // Every machine that's displaying the scoreboard helpfully stores the last 20 values of
    // points.json for us, in case of catastrophe. Thanks, y'all!
    //
    // We have been doing some variation on this "everybody backs up the server state" trick since 2009.
    // We have needed it 0 times.
    let stateHistory = JSON.parse(localStorage.getItem("stateHistory")) || []
    if (stateHistory.length >= 20) {
      stateHistory.shift()
    }
    stateHistory.push(state)
    localStorage.setItem("stateHistory", JSON.stringify(stateHistory))

    draw();
  } 

  function draw() {
    let teams = {}
    let categories = {} // map[string][team]int
    let highestCategoryScore = {} // map[string]int

    let uiRanking=cortex.ranking.valueOf()
    let element = document.getElementById("rankings")


    // Initialize data structures
    for (let teamId in teamNames) {
      teams[teamId] = {
        categoryScore: {},        // map[string]int
        trackScore: {},           // map[string]int
        overallScore: 0,          // int
        historyLine: [],          // []{x: int, y: int}
        history: [],              // []{t: timestamp, c: category, s: int}
        name: teamNames[teamId],
        id: teamId
      }
    }

    // Dole out points
    pointsLog.sort(MOTH_COMP_POINTSLOGTIME)

    for (let entry of pointsLog) {
      let timestamp = entry[0]
      let teamId = entry[1]
      let category = entry[2]
      let points = parseInt(entry[3])
      
      let team = teams[teamId]

      let score = team.categoryScore[category] || 0
      let trackScore = team.trackScore[trackMap[category]] || 0
      score += points
      trackScore += points

      team.categoryScore[category] = score

      team.history.push({t: new Date(timestamp * 1000), cat: category, score: points})

      if (!categories[category]) {
        categories[category]={}
      }

      categories[category][teamId]=score
    }
  
    // Search Team score aggregates for highest scores and key markers
    for (let cat in categories) {
      let scores=Object.values(categories[cat])
      scores.sort(MOTH_COMP_SCORE)
      scores.reverse();
      let highest=scores[0]
      highestCategoryScore[cat]=highest.valueOf()
    }

    // Compute overall scores based on current highest
    for (let teamId in teams) {
      let team = teams[teamId]
      team.overallScore = 0
      for (let cat in team.categoryScore) {
        team.overallScore += team.categoryScore[cat] / highestCategoryScore[cat]
      }

      // HistoryLine
      let overall = 0
      for (let history in team.history) {
        let entry=team.history[history]
        overall+=entry.score/highestCategoryScore[entry.cat]
        team.historyLine.push({t: entry.t, y: overall.toFixed(2)})
      }

    }

  
    // Figure out how to order each team on the scoreboard
    let winners = []
    for (let teamId in teams) {
      winners.push(teams[teamId])
    }
    winners.sort(cortex.sorting)
    winners.reverse()
    
    // Let's make some better names for things we've computed
    let winningScore = winners[0].overallScore
    let numCategories = Object.keys(highestCategoryScore).length
  
    // Clear out the element we're about to populate
    Array.from(element.childNodes).map(e => e.remove())
  
    let maxWidth = (100 / winningScore)
    let avgWidth = (100 / numCategories)

    // Pre-load heading
    let headingRow=document.createElement("div")
    headingRow.id="rankHeading";
    let headingRowPoints=document.createElement("span")
    headingRowPoints.classList.add("teampoints")
    headingRowPoints.classList.add("inv")

    let headingNcat=0
    for (let category in highestCategoryScore) {
      let bar=document.createElement("span")
      bar.title=category
      bar.style.width=avgWidth +"%"
      bar.classList.add("cat" +headingNcat)
      bar.textContent=category
      bar.dataset.category=category;
      bar.onclick=sortByCategory;
      
      headingRowPoints.appendChild(bar);
      headingNcat+=1
    }

    headingRow.appendChild(headingRowPoints);
    element.appendChild(headingRow)

    for (let team of winners) {
      let row = document.createElement("div")
      row.classList.add("team");
      row.dataset.overallScore=team.overallScore.toFixed(2);

      let ncat = 0

      let teamPoints=document.createElement("span")
      teamPoints.classList.add("teampoints")

      let leader=[];

      for (let category in highestCategoryScore) {
        let catHigh = highestCategoryScore[category]
        let catTeam = team.categoryScore[category] || 0
        let catPct = catTeam / catHigh
        let width = (maxWidth * catPct)
        let catWidth = (avgWidth * catPct)
        
        let bar = document.createElement("span")
        bar.classList.add("category")
        bar.classList.add("cat" + ncat)
        bar.dataset.standardWidth=width;
        bar.dataset.categoryWidth=catWidth;
        bar.dataset.category=category;
        bar.dataset.points=catTeam;
        bar.dataset.categoryMargin=(avgWidth - catWidth)
        bar.title = bar.dataset.category + ": " + bar.dataset.points

        if ((catTeam ==  catHigh) && (trackMap[category])){
          leader.push(trackMap[category])
        }

        displayMothRanking(uiRanking, bar);
  
        teamPoints.appendChild(bar)
        ncat += 1
      }

      row.appendChild(teamPoints)
  
      let te = document.createElement("span")
      te.classList.add("teamname")
      te.textContent = team.name

      for (let track in leader) {
        te.classList.add("leader");

        let img=document.createElement("img");
        img.classList.add("icon");
        img.classList.add("track-"+leader[track]);

        te.prepend(img);
      }

      row.appendChild(te)
  
      element.appendChild(row)
    }
    
    let datasets = []
    for (let i in winners) {
      if (i > 5) {
        break
      }
      let team = winners[i]
      let color = chartColors[i % chartColors.length]
      datasets.push({
        label: team.name,
        backgroundColor: color,
        borderColor: color,
        data: team.historyLine,
        lineTension: 0,
        fill: false
      })
    }
    let config = {
      type: "line",
      data: {
        datasets: datasets
      },
      options: {
        responsive: true,
        scales: {
          xAxes: [{
            display: true,
            type: "time",
            time: {
              tooltipFormat: "ll HH:mm"
            },
            scaleLabel: {
              display: true,
              labelString: "Time"
            }
          }],
          yAxes: [{
            display: true,
            scaleLabel: {
              display: true,
              labelString: "Points"
            }
          }]
        },
        tooltips: {
          mode: "index",
          intersect: false
        },
        hover: {
          mode: "nearest",
          intersect: true
        }
      }
    }
    
    let chart = document.querySelector("#chart")
    if (chart) {
      let canvas = chart.querySelector("canvas")
      if (! canvas) {
        canvas = document.createElement("canvas")
        chart.appendChild(canvas)
      }
      
      let myline = new Chart(canvas.getContext("2d"), config)
      myline.update()
    }
  }
  
  function refresh() {
    fetch("state")
    .then(resp => {
      return resp.json()
    })
    .then(obj => {
      update(obj)
    })
    .catch(err => {
      console.log(err)
    })
  }

  let imgRL=null;
  let imgLR=null;
  let canvas=null;

  function setScreenSaver() {
    let uiScreenSaver=document.querySelector("#mothScreenSaver");

    if (uiScreenSaver.checked) {
      if (canvas == null) {
        canvas=document.createElement("picture");
        canvas.id="gibson"
      }

      document.querySelector("body").appendChild(canvas);

      if (imgRL == null) {
        imgRL=document.createElement("img");
        imgRL.classList.add("moth");
        imgRL.src="luna-moth.svg";
        imgRL.style.animation="slideRL 4s linear infinite";
        imgRL.dataset.direction="RL";
        imgRL.addEventListener("animationiteration", lunaRLIterListener, false);
      }

      imgRL.style.top="4em";

      if (imgLR == null) {
        imgLR=document.createElement("img");
        imgLR.classList.add("moth");
        imgLR.src="luna-moth.svg";
        imgLR.style.animation="slideLR 4s linear infinite";
        imgLR.dataset.direction="LR";
        imgLR.addEventListener("animationiteration", lunaLRIterListener, false);
      }

      imgLR.style.top=0;

      canvas.appendChild(imgLR);
      canvas.appendChild(imgRL);

      setScreenSaverBg();

      setTimeout(lunaShadow, Math.random()*100+400);

    } else {
      canvas=document.querySelector("#gibson");
      if (canvas) {
        canvas.remove();
      }

      canvas=null;
      imgRL=null;
      imgLR=null;
    }
  }

  function setScreenSaverBg() {
    let options=["svg", "window", "combined"];
    document.querySelectorAll("input[name=mothScreenSaverBg]").forEach(function(item) {
      if (item.checked === true) {
        cortex.screenSaverBg=item.value;
      }
    });


    if (canvas !== null) {
      for (let i=0; i<options.length; i++) {
        if (options[i] != cortex.screenSaverBg) {
          canvas.classList.remove(options[i]);
        }

        canvas.classList.add(cortex.screenSaverBg);
      }
    }
  }


  function animateScreenSaver() {
    let img=document.querySelector("#screensaverImg");

    if (img == null) {
      return false;
    }

    let coord=img.getBoundingClientRect();

    img.style.transitionDuration="0s"
    img.style.transform="translate("+ Math.max((-1*coord.width), (-1*(coord.width+coord.x))) + "px, " + (coord.y + coord.height) + "px) rotate(135deg)";
       
    setTimeout(function() {
      img.style.transitionDuration="5s"
      img.style.transform="translate(100vw, " + (coord.y + coord.height) + "px) rotate(135deg)";
    }, 100);

    setTimeout(animateScreenSaver, 5200);

    return true;
  }

  function setShowLeaderIcons() {
    let uiShowLeaderIcons=document.querySelector("#mothLeaderIcons");
    let body=document.querySelector("body");

    (uiShowLeaderIcons.checked)? body.classList.add("fun") : body.classList.remove("fun");

  }

  function setRankingTrack() {
    let uiRankingType=document.querySelector("input[name=rankingPerspective]:checked");

    switch (uiRankingType.value) {
      case "track":
        cortex.ranking=MOTH_RANKING_TRACK;
        break;

      case "category":
        cortex.ranking=MOTH_RANKING_CATEGORY;
        break;

      case "standard":
      default:
        cortex.ranking=MOTH_RANKING_STANDARD;
    }

    let uiRanking=cortex.ranking.valueOf();

    document.querySelectorAll(".category").forEach(function(item) {
      displayMothRanking(uiRanking, item)
    });

    switch (uiRanking) {
      case MOTH_RANKING_CATEGORY:
      case MOTH_RANKING_TRACK:
        document.querySelector("#rankings").classList.add("track");
        break;

      case MOTH_RANKING_STANDARD:
      default:
        cortex.sorting = MOTH_COMP_TEAMOVERALL;
        setTimeout(sortByCategory, 1000);
        document.querySelector("#rankings").classList.remove("track");
    }
  }

  function displayMothRanking(rankingStyle, obj) {
      switch (rankingStyle) {
        case MOTH_RANKING_CATEGORY:
        case MOTH_RANKING_TRACK:
          obj.style.width=obj.dataset.categoryWidth + "%";
          obj.style.marginRight=obj.dataset.categoryMargin + "%";
          obj.textContent = obj.dataset.points;
          break;

        case MOTH_RANKING_STANDARD:
        default:
          obj.style.width=obj.dataset.standardWidth + "%";
          obj.style.marginRight=0+"%";
          obj.textContent = obj.dataset.category + ": " + obj.dataset.points

     }
  }

  // Sort winners by category score
  function sortByCategory() {
    let cat=(this.dataset)? this.dataset.category || MOTH_COMP_TEAMOVERALL : MOTH_COMP_TEAMOVERALL;

    let teamOrder=Array.from(document.querySelectorAll("#rankings div.team"));
    let rankings=document.querySelector("#rankings");

    // Grab current screen positions of objects
    let coords=[]

    for (let i=0; i<teamOrder.length; i++) {
      coords.push({ x: teamOrder[i].offsetLeft, y: teamOrder[i].offsetTop });
    }

//    console.log(teamOrder[0].offsetTop);

//    teamOrder.forEach(function(item) {
//      item.style.top=item.offsetTop +"px";
//      item.style.left=item.offsetLeft +"px";
//    });

    if (cat == MOTH_COMP_TEAMOVERALL) {
      teamOrder.sort(function(a, b) {
        return b.dataset.overallScore - a.dataset.overallScore;
      });

    } else {
      teamOrder.sort(function(a, b) {
        let aa=a.querySelector("span[data-category=" + cat + "]") || { dataset : {points: 0}};
        let bb=b.querySelector("span[data-category=" + cat + "]") || { dataset : {points: 0}};

        return bb.dataset.points - aa.dataset.points;
      });
    }

    // Move elements without rearranging the DOM
    // Seems to be a couple ways to do this.
    //  1. appendChild in DOM will move the element without transition
    //  2. translate to move the elements where they should be with transition
    teamOrder.forEach(function(item, i) {
      item.style.transform="translate(" + (coords[i]["x"] - item.offsetLeft) + "px, " + (coords[i]["y"] - item.offsetTop) +"px)"


//      rankings.appendChild(item);
//      item.style.transform="translate(" + (item-offsetLeft - coords[i]["x"]) + "px, " + (item.offsetTop - coords[i]["y"]) +"px)"

    });
  }

  /* Luna Moth ScreenSaver */
  function lunaRLIterListener(event) {
    let coord=imgRL.getBoundingClientRect();

    if ((coord.y + coord.height) > canvas.clientHeight) {
      // Repeat in infinite loop
//      imgRL.style.top="4em";

      // Uncomment to run through screen once
      imgRL.style.animationPlayState="paused";
    } else {
      imgRL.style.top=(coord.y + (2*coord.height))+"px";
    }
  }

  function lunaLRIterListener(event) {
    let coord=imgLR.getBoundingClientRect();

    if ((coord.y + coord.height) > canvas.clientHeight) {
      // Repeat in infinite loop
//      imgLR.style.top="0px";

      // Uncomment to run through screen once
      imgLR.style.animationPlayState="paused";
    } else {
      imgLR.style.top=(coord.y + (2*coord.height))+"px";
    }
  }

  function lunaShadow() {
  //  let coordCanvas=canvas.getBoundingClientRect();
    let targets=[imgRL, imgLR];

    targets.forEach(function(item) {
      if (item == null) {
        return false;
      }

      let coord=item.getBoundingClientRect();
      let div=document.createElement("div");
      div.className="shadow "+item.dataset.direction;
      div.style.top=item.style.top;
      div.style.left=coord.left +"px";
 
      canvas.appendChild(div);
    });

    // Only run through screen once
    if ((imgRL != null) && (imgRL.style.animationPlayState != "paused")) {
      setTimeout(lunaShadow, Math.random()*400+200);
    }

  }


  function init() {
    let base = window.location.href.replace("scoreboard.html", "")
    let location = document.querySelector("#location")
    let params = new URLSearchParams(document.location.search.substring(1));

    if (location) {
      location.textContent = base
    }

    // Grab initial settings and set event handlers
    document.querySelectorAll("input[name=rankingPerspective]").forEach(function(item) {
      if (item.value == params.get(item.name)) {
        item.checked=true;
      }
      item.onchange=setRankingTrack;
    });;

    // Leader Icons
    let leaderIcons=document.querySelector("#mothLeaderIcons");
    if (leaderIcons) {
      leaderIcons.checked=(params.get(leaderIcons.name) == leaderIcons.value);
      leaderIcons.onchange=setShowLeaderIcons;
    }

    // Screensaver
    let screenSaver=document.querySelector("#mothScreenSaver");
    if (screenSaver) {
      screenSaver.checked=(params.get(screenSaver.name) == screenSaver.value);
      screenSaver.onchange=setScreenSaver;
    }

    document.querySelectorAll("input[name=mothScreenSaverBg]").forEach(function(item) {
      if (item.value == params.get(item.name)) {
        item.checked=true;
      }

      item.onchange=setScreenSaverBg;;
    });

    setRankingTrack();
    setShowLeaderIcons();
    setScreenSaver();

    setInterval(refresh, 60000)
    refresh()
  }



  init()
}


if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", scoreboardInit)
} else {
  scoreboardInit()
}


