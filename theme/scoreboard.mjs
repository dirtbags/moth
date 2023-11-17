import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

const server = new moth.Server(".")
/** Don't let any team's score exceed this percentage width */
const MaxScoreWidth = 90

/**
 * Returns a promise that resolves after timeout.
 *
 * This uses setTimeout instead of some other fancy thing like
 * requestAnimationFrame, because who actually cares about scoreboard update
 * framerate?
 *
 * @param {Number} timeout How long to sleep (milliseconds)
 * @returns {Promise}
 */
function sleep(timeout) {
  return new Promise(resolve => setTimeout(resolve, timeout));
}

/**
 * Pull new points log, and update the scoreboard.
 * 
 * The update is animated, because I think that looks cool.
 */
async function update() {
  let config = {}
  try {
    config = await common.Config()
  }
  catch (err) {
    console.warn("Parsing config.json:", err)
  }

  // Pull configuration settings
  if (!config.Scoreboard) {
    console.warn("config.json has empty Scoreboard section")
  }
  let ScoreboardConfig = config.Scoreboard ?? {}
  let state = await server.GetState()

  // Show URL of server
  for (let e of document.querySelectorAll(".location")) {
    e.textContent = common.BaseURL
    e.classList.toggle("hidden", !(ScoreboardConfig.DisplayServerURLWhenEnabled && state.Enabled))
  }

  // Rotate views
  for (let e of document.querySelectorAll(".rotate")) {
    e.appendChild(e.firstChild)
  }

  // Render rankings
  for (let e of document.querySelectorAll(".rankings")) {
    if (e.classList.contains("classic")) {
      classicRankings(e, state, ScoreboardConfig)
    } else if (e.classList.contains("category")) {
      categoryRankings(e, state, ScoreboardConfig)
    }
  }

}

async function classicRankings(rankingsElement, state, ScoreboardConfig) {
  let ReplayHistory = ScoreboardConfig.ReplayHistory ?? false
  let ReplayDurationMS = ScoreboardConfig.ReplayDurationMS ?? 300
  let ReplayFPS = ScoreboardConfig.ReplayFPS ?? 24

  let logSize = state.PointsLog.length

  // Figure out the timing so that we can replay the scoreboard in about
  // ReplayDurationMS.
  let frameModulo = 1
  let delay = 0
  while (delay < (common.Second / ReplayFPS)) {
    frameModulo += 1
    delay = ReplayDurationMS / (logSize / frameModulo)
  }

  let frame = 0
  for (let scores of state.ScoresHistory()) {
    frame += 1
    if (frame < state.PointsLog.length) { // Always render the last frame
      if (!ReplayHistory || (frame % frameModulo)) { // Skip if we're not animating, or if we need to drop frames
        continue
      }
    }

    while (rankingsElement.firstChild) rankingsElement.firstChild.remove()

    let sortedTeamIDs = [...scores.TeamIDs]
    sortedTeamIDs.sort((a, b) => scores.CyFiScore(a) - scores.CyFiScore(b))
    sortedTeamIDs.reverse()
    
    let topScore = scores.CyFiScore(sortedTeamIDs[0])
    for (let teamID of sortedTeamIDs) {
      let teamName = state.TeamNames[teamID] ?? "rodney"
      
      let row = rankingsElement.appendChild(document.createElement("div"))
      
      let teamname = row.appendChild(document.createElement("span"))
      teamname.textContent = teamName
      teamname.classList.add("teamname")
      
      let teampoints = row.appendChild(document.createElement("span"))
      teampoints.classList.add("teampoints")
      for (let category of scores.Categories) {
        let score = scores.CyFiCategoryScore(category, teamID)
        if (!score) {
          continue
        }

        // XXX: Figure out how to do this properly with flexbox
        let block = row.appendChild(document.createElement("span"))
        let points = scores.GetPoints(category, teamID)
        let width = MaxScoreWidth * score / topScore
        let categoryNumber = [...scores.Categories].indexOf(category)

        block.textContent = category
        block.title = `${points} points`
        block.style.width = `${width}%`
        block.classList.add("category", `cat${categoryNumber}`)
        block.classList.toggle("topscore", (score == 1) && ScoreboardConfig.ShowCategoryLeaders)

        categoryNumber += 1
      } 
    }
    await sleep(delay)
  }

  for (let e of document.querySelectorAll(".no-scores")) {
    e.innerHTML = ScoreboardConfig.NoScoresHtml
    e.classList.toggle("hidden", frame > 0)
  }
}

/**
 * 
 * @param {*} rankingsElement 
 * @param {moth.State} state 
 * @param {*} ScoreboardConfig 
 */
async function categoryRankings(rankingsElement, state, ScoreboardConfig) {
  while (rankingsElement.firstChild) rankingsElement.firstChild.remove()
  let scores = state.CurrentScores()
  for (let category of scores.Categories) {
    let categoryBox = rankingsElement.appendChild(document.createElement("div"))
    categoryBox.classList.add("category")

    categoryBox.appendChild(document.createElement("h2")).textContent = category

    let categoryScores = []
    for (let teamID in state.TeamNames) {
      categoryScores.push({
        teamName: state.TeamNames[teamID],
        score: scores.GetPoints(category, teamID),
      })
    }
    categoryScores.sort((a, b) => b.score - a.score)
    
    let table = categoryBox.appendChild(document.createElement("table"))
    let rows = 0
    for (let categoryScore of categoryScores) {
      let row = table.appendChild(document.createElement("tr"))
      row.appendChild(document.createElement("td")).textContent = categoryScore.teamName
      let td = row.appendChild(document.createElement("td"))
      td.textContent = categoryScore.score
      td.classList.add("number")
      rows += 1
      if (rows == 5) {
        break
      }
    }
  }
}

function init() {
  setInterval(update, common.Minute)
  update()
}

common.WhenDOMLoaded(init)