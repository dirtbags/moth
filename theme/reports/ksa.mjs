import * as moth from "../moth.mjs"
import * as common from "../common.mjs"

const server = new moth.Server("../")

/**
 * Update "doing" indicators
 * 
 * @param {String | null} what Text to display, or null to not update text
 * @param {Number | null} finished Percentage complete to display, or null to not update progress
 */
function doing(what, finished = null) {
    for (let e of document.querySelectorAll(".doing")) {
        e.classList.remove("hidden")
        if (what) {
            e.textContent = what
        }
        if (finished) {
            e.value = finished
        } else {
            e.removeAttribute("value")
        }
    }
}
function done() {
    for (let e of document.querySelectorAll(".doing")) {
        e.classList.add("hidden")
    }
}

async function GetNice() {
    let NiceElementsByIdentifier = {}
    let resp = await fetch("NICEFramework2017.json")
    let obj = await resp.json()
    for (let e of obj.elements) {
        NiceElementsByIdentifier[e.element_identifier] = e
    }
    return NiceElementsByIdentifier
}

/**
 * Fetch a puzzle, and fill its KSAs and rows.
 *
 * This is done once per puzzle, in an asynchronous function, allowing the
 * application to perform multiple blocking operations simultaneously.
 */
async function FetchAndFill(puzzle, KSAs, rows) {
    try {
        await puzzle.Populate()
    }
    catch (error) {
        // Keep on going with whatever Populate was able to fill
    }
    for (let KSA of (puzzle.KSAs || [])) {
        KSAs.add(KSA)
    }

    for (let row of rows) {
        row.querySelector(".category").textContent = puzzle.Category
        row.querySelector(".points").textContent = puzzle.Points
        row.querySelector(".ksas").textContent = (puzzle.KSAs || []).join(" ")
        row.querySelector(".error").textContent = puzzle.Error.Body
    }
}

async function init() {
    doing("Fetching NICE framework data")
    let nicePromise = GetNice()

    doing("Retrieving server state")
    let state = await server.GetState()

    doing("Retrieving all puzzles")
    let KSAsByCategory = {}
    let puzzlerowTemplate = document.querySelector("template#puzzlerow")
    let puzzles = state.Puzzles()
    let promises = []
    for (let category of state.Categories()) {
        KSAsByCategory[category] = new Set()
    }
    let pending = puzzles.length
    for (let puzzle of puzzles) {
        // Make space in the table, so everything fills in sorted order
        let rows = []
        for (let tbody of document.querySelectorAll("tbody")) {
            let row = puzzlerowTemplate.content.cloneNode(true).firstElementChild
            tbody.appendChild(row)
            rows.push(row)
        }

        // Queue up a fetch, and update progress bar
        let promise = FetchAndFill(puzzle, KSAsByCategory[puzzle.Category], rows)
        promises.push(promise)
        promise.then(() => doing(null, 1 - (--pending / puzzles.length)))

        if (promises.length > 50) {
            // Chrome runs out of resources if you queue up too many of these at once
            await Promise.all(promises)
            promises = []
        }
    }
    await Promise.all(promises)

    doing("Retrieving NICE identifiers")
    let NiceElementsByIdentifier = await nicePromise


    doing("Filling KSAs By Category")
    let allKSAs = new Set()
    for (let div of document.querySelectorAll(".KSAsByCategory")) {
        for (let category of state.Categories()) {
            doing(`Filling KSAs for category: ${category}`)
            let KSAs = [...KSAsByCategory[category]]
            KSAs.sort()

            div.appendChild(document.createElement("h3")).textContent = category
            let ul = div.appendChild(document.createElement("ul"))
            for (let k of KSAs) {
                let ksa = k.split(/\s+/)[0]
                let ne = NiceElementsByIdentifier[ksa] || { text: "???" }
                let text = `${ksa}: ${ne.text}`
                ul.appendChild(document.createElement("li")).textContent = text
                allKSAs.add(text)
            }
        }
    }

    doing("Filling KSAs")
    for (let e of document.querySelectorAll(".allKSAs")) {
        let KSAs = [...allKSAs]
        KSAs.sort()
        for (let text of KSAs) {
            e.appendChild(document.createElement("li")).textContent = text
        }
    }

    done()
}

common.WhenDOMLoaded(init)
