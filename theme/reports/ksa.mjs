import * as moth from "../moth.mjs"

function doing(what) {
    for (let e of document.querySelectorAll(".doing")) {
        if (what) {
            e.style.display = "inherit"
        } else {
            e.style.display = "none"
        }
        for (let p of e.querySelectorAll("p")) {
            p.textContent = what
        }
    }
}

async function init() {
    let server = new moth.Server("../")

    doing("Retrieving server state")
    let state = await server.GetState()

    doing("Retrieving all puzzles")
    let puzzles = state.Puzzles()
    for (let p of puzzles) {
        await p.Populate().catch(x => {})
    }

    doing("Filling table")
    let puzzlerowTemplate = document.querySelector("template#puzzlerow")
    for (let tbody of document.querySelectorAll("tbody")) {
        for (let puzzle of puzzles) {
            let row = puzzlerowTemplate.content.cloneNode(true)
            row.querySelector(".category").textContent = puzzle.Category
            row.querySelector(".points").textContent = puzzle.Points
            row.querySelector(".ksas").textContent = puzzle.KSAs.join(" ")
            row.querySelector(".error").textContent = puzzle.Error.Body
            tbody.appendChild(row)
        }
    }

    doing()
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
  