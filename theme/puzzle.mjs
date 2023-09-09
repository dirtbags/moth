import * as moth from "./moth.mjs"

function puzzleElement(clear=true) {
    let e = document.querySelector("#puzzle")
    if (clear) {
        while (e.firstChild) e.firstChild.remove()
    }
    return e
}

function error(message) {
    let e = puzzleElement().appendChild(document.createElement("p"))
    e.classList.add("error")
    e.textContent = message
}

async function loadPuzzle(category, points) {
    let server = new moth.Server()
    let puzzle = server.GetPuzzle(category, points)
    await puzzle.Populate()

    let title = `${category} ${points}`
    document.querySelector("title").textContent = title
    document.querySelector("#title").textContent = title
    document.querySelector("#authors").textContent = puzzle.Authors.join(", ")
    puzzleElement().innerHTML = puzzle.Body
}

function hashchange() {
    // Tell user we're loading
    puzzleElement().appendChild(document.createElement("progress"))
    for (let qs of ["#authors", "#title", "title"]) {
        for (let e of document.querySelectorAll(qs)) {
            e.textContent = "[loading]"
        }
    }

    let hashpart = location.hash.split("#")[1] || ""
    let catpoints = hashpart.split(":")
    let category = catpoints[0]
    let points = Number(catpoints[1])
    if (!category && !points) {
        error(`Doesn't look like a puzzle reference: ${hashpart}`)
        return
    }

    loadPuzzle(category, points)
    .catch(err => error(err))
}

function init() {
    window.addEventListener("hashchange", hashchange)
    hashchange()
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
