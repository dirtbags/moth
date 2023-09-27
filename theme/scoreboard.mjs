import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

const server = new moth.Server(".")
const ReplayDuration = 0.3 * common.Second
const MaxFrameRate = 60
/** Don't let any team's score exceed this percentage width */
const MaxScoreWidth = 95

/**
 * Returns a promise that resolves after timeout.
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
  let config = await common.Config()
  for (let e of document.querySelectorAll(".location")) {
    e.textContent = common.BaseURL
    e.classList.toggle("hidden", !config.URLInScoreboard)
  }

  let state = await server.GetState()
  let rankingsElement = document.querySelector("#rankings")
  let logSize = state.PointsLog.length

  // Figure out the timing so that we can replay the scoreboard in about
  // ReplayDuration, but no more than 24 frames per second.
  let frameModulo = 1
  let delay = 0
  while (delay < (common.Second / MaxFrameRate)) {
    frameModulo += 1
    delay = ReplayDuration / (logSize / frameModulo)
  }

  let frame = 0
  for (let scores of state.ScoresHistory()) {
    frame += 1
    if ((frame < state.PointsLog.length) && (frame % frameModulo)) {
      continue
    }

    while (rankingsElement.firstChild) rankingsElement.firstChild.remove()

    let sortedTeamIDs = [...scores.TeamIDs]
    sortedTeamIDs.sort((a, b) => scores.CyFiScore(a) - scores.CyFiScore(b))
    sortedTeamIDs.reverse()
    
    let topScore = scores.CyFiScore(sortedTeamIDs[0])
    for (let teamID of sortedTeamIDs) {
      let teamName = state.TeamNames[teamID]
      
      let row = rankingsElement.appendChild(document.createElement("div"))
      
      let teamname = row.appendChild(document.createElement("span"))
      teamname.textContent = teamName
      teamname.classList.add("teamname")
      
      let categoryNumber = 0
      let teampoints = row.appendChild(document.createElement("span"))
      teampoints.classList.add("teampoints")
      for (let category of scores.Categories) {
        let score = scores.CyFiCategoryScore(category, teamID)
        if (!score) {
          continue
        }

        let block = teampoints.appendChild(document.createElement("span"))
        let points = scores.GetPoints(category, teamID)
        let width = MaxScoreWidth * score / topScore

        block.textContent = category
        block.title = `${points} points`
        block.style.width = `${width}%`
        block.classList.add("category", `cat${categoryNumber}`)

        categoryNumber += 1
      } 
    }
    await sleep(delay)
  }
}

function init() {
  setInterval(update, common.Minute)
  update()
}

common.WhenDOMLoaded(init)