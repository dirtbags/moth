// jshint asi:true

function scoreboardInit() {
  
  chartColors = [
    "rgb(255, 99, 132)",
    "rgb(255, 159, 64)",
    "rgb(255, 205, 86)",
    "rgb(75, 192, 192)",
    "rgb(54, 162, 235)",
    "rgb(153, 102, 255)",
    "rgb(201, 203, 207)"
  ]
  
  function update(state) {
    for (let rotate of document.querySelectorAll(".rotate")) {
      rotate.appendChild(rotate.firstElementChild)
    }
    
    let element = document.getElementById("rankings")
    let teamNames = state.teams
    let pointsLog = state.points
  
    // Every machine that's displaying the scoreboard helpfully stores the last 20 values of
    // points.json for us, in case of catastrophe. Thanks, y'all!
    //
    // We have been doing some variation on this "everybody backs up the server state" trick since 2009.
    // We have needed it 0 times.
    let pointsHistory = JSON.parse(localStorage.getItem("pointsHistory")) || []
    if (pointsHistory.length >= 20) {
      pointsHistory.shift()
    }
    pointsHistory.push(pointsLog)
    localStorage.setItem("pointsHistory", JSON.stringify(pointsHistory))
  
    let teams = {}
    let highestCategoryScore = {} // map[string]int
  
    // Initialize data structures
    for (let teamId in teamNames) {
        teams[teamId] = {
        categoryScore: {},        // map[string]int
        overallScore: 0,          // int
        historyLine: [],          // []{x: int, y: int}
        name: teamNames[teamId],
        id: teamId
      }
    }
  
    // Dole out points
    for (let entry of pointsLog) {
      let timestamp = entry[0]
      let teamId = entry[1]
      let category = entry[2]
      let points = entry[3]
      
      let team = teams[teamId]
  
      let score = team.categoryScore[category] || 0
      score += points
      team.categoryScore[category] = score
  
      let highest = highestCategoryScore[category] || 0
      if (score > highest) {
        highestCategoryScore[category] = score
      }
    }

    for (let teamId in teamNames) {
      teams[teamId].categoryScore = {}
    }
    
    for (let entry of pointsLog) {
      let timestamp = entry[0]
      let teamId = entry[1]
      let category = entry[2]
      let points = entry[3]
      
      let team = teams[teamId]
  
      let score = team.categoryScore[category] || 0
      score += points
      team.categoryScore[category] = score
  
      let overall = 0
      for (let cat in team.categoryScore) {
        overall += team.categoryScore[cat] / highestCategoryScore[cat]
      }
  
      team.historyLine.push({t: new Date(timestamp  * 1000), y: overall})
    }
  
    // Compute overall scores based on current highest
    for (let teamId in teams) {
      let team = teams[teamId]
      team.overallScore = 0
      for (let cat in team.categoryScore) {
        team.overallScore += team.categoryScore[cat] / highestCategoryScore[cat]
      }
    }
  
    // Sort by team score
    function teamCompare(a, b) {
      return a.overallScore - b.overallScore
    }
  
    // Figure out how to order each team on the scoreboard
    let winners = []
    for (let teamId in teams) {
      winners.push(teams[teamId])
    }
    winners.sort(teamCompare)
    winners.reverse()
    
    // Let's make some better names for things we've computed
    let winningScore = winners[0].overallScore
    let numCategories = Object.keys(highestCategoryScore).length
  
    // Clear out the element we're about to populate
    Array.from(element.childNodes).map(e => e.remove())
  
    let maxWidth = 100 / winningScore
    for (let team of winners) {
      let row = document.createElement("div")
      let ncat = 0
      for (let category in highestCategoryScore) {
        let catHigh = highestCategoryScore[category]
        let catTeam = team.categoryScore[category] || 0
        let catPct = catTeam / catHigh
        let width = maxWidth * catPct
        
        let bar = document.createElement("span")
        bar.classList.add("category")
        bar.classList.add("cat" + ncat)
        bar.style.width = width + "%"
        bar.textContent = category + ": " + catTeam
        bar.title = bar.textContent
  
        row.appendChild(bar)
        ncat += 1
      }
  
      let te = document.createElement("span")
      te.classList.add("teamname")
      te.textContent = team.name
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
    fetch("points.json")
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

  function init() {
    let base = window.location.href.replace("scoreboard.html", "")
    let location = document.querySelector("#location")
    if (location) {
      location.textContent = base
    }
  
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
