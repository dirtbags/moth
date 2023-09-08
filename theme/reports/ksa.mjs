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
    let puzzlerowTemplate = document.querySelector("template#puzzlerow")
    let puzzles = state.Puzzles()
    for (let puzzle of puzzles) {
        await puzzle.Populate().catch(x => {})
    }

    doing("Filling tables")
    let KSAsByCategory = {}
    for (let puzzle of puzzles) {
        let KSAs = KSAsByCategory[puzzle.Category]
        if (!KSAs) {
            KSAs = new Set()
            KSAsByCategory[puzzle.Category] = KSAs
        }
        for (let KSA of (puzzle.KSAs || [])) {
            KSAs.add(KSA)
        }

        for (let tbody of document.querySelectorAll("tbody")) {
            let row = puzzlerowTemplate.content.cloneNode(true)
            row.querySelector(".category").textContent = puzzle.Category
            row.querySelector(".points").textContent = puzzle.Points
            row.querySelector(".ksas").textContent = (puzzle.KSAs || []).join(" ")
            row.querySelector(".error").textContent = puzzle.Error.Body
            tbody.appendChild(row)
        }
    }

    doing("Filling KSAs By Category")
    for (let div of document.querySelectorAll(".KSAsByCategory")) {
        for (let category of state.Categories()) {
            let KSAs = [...KSAsByCategory[category]]
            KSAs.sort()

            div.appendChild(document.createElement("h3")).textContent = category
            let ul = div.appendChild(document.createElement("ul"))
            for (let k of KSAs) {
                ul.appendChild(document.createElement("li")).textContent = k
            }
        }
    }

    doing()
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
  