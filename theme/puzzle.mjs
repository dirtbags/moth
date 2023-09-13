import * as moth from "./moth.mjs"

/**
 * Handle a submit event on a form.
 * 
 * This event will be called when the user submits the form,
 * either by clicking a "submit" button,
 * or by some other means provided by the browser,
 * like hitting the Enter key.
 * 
 * @param {Event} event 
 */
function formSubmitHandler(event) {
    event.preventDefault()
    console.log(event)
}

/**
 * Handle an input event on the answer field.
 * 
 * @param {Event} event 
 */
async function answerInputHandler(event) {
    let answer = event.target.value
    let correct = await window.app.puzzle.IsPossiblyCorrect(answer)
    for (let ok of event.target.parentElement.querySelectorAll(".answer_ok")) {
        if (correct) {
            ok.textContent = "⭕"
            ok.title = "Possibly correct"
        } else {
            ok.textContent = "❌"
            ok.title = "Definitely not correct"
        }
    }
}

/**
 * Return the puzzle content element, possibly with everything cleared out of it.
 * 
 * @param {Boolean} clear Should the element be cleared of children? Default true.
 * @returns {Element}
 */
function puzzleElement(clear=true) {
    let e = document.querySelector("#puzzle")
    if (clear) {
        while (e.firstChild) e.firstChild.remove()
    }
    return e
}

/**
 * Display an error in the puzzle area, and also send it to the console.
 * 
 * This makes it so the user can see a bit more about what the problem is.
 * 
 * @param {String} error 
 */
function error(error) {
    console.error(error)
    let e = puzzleElement().appendChild(document.createElement("pre"))
    e.classList.add("error")
    e.textContent = error.Body || error
}

/**
 * Set the answer and invoke input handlers.
 * 
 * This makes sure the Circle Of Success gets updated.
 * 
 * @param {String} s 
 */
function setanswer(s) {
    let e = document.querySelector("#answer")
    e.value = s
    e.dispatchEvent(new Event("input"))
}

/**
 * Load the given puzzle.
 * 
 * @param {String} category 
 * @param {Number} points 
 */
async function loadPuzzle(category, points) {
    console.group("Loading puzzle:", category, points)
    let contentBase = new URL(`content/${category}/${points}/`, location)
    
    // Tell user we're loading
    puzzleElement().appendChild(document.createElement("progress"))
    for (let qs of ["#authors", "#title", "title"]) {
        for (let e of document.querySelectorAll(qs)) {
            e.textContent = "[loading]"
        }
    }    

    let server = new moth.Server()
    let puzzle = server.GetPuzzle(category, points)
    console.time("Populate")
    await puzzle.Populate()
    console.timeEnd("Populate")

    let title = `${category} ${points}`
    document.querySelector("title").textContent = title
    document.querySelector("#title").textContent = title
    document.querySelector("#authors").textContent = puzzle.Authors.join(", ")
    puzzleElement().innerHTML = puzzle.Body

    console.info("Adding attached scripts...")
    for (let script of (puzzle.Scripts || [])) {
        let st = document.createElement("script")
        document.head.appendChild(st)
        st.src = new URL(script, contentBase)
      }

    console.info("Listing attached files...")
    for (let fn of (puzzle.Attachments || [])) {
        let li = document.createElement("li")
        let a = document.createElement("a")
        a.href = new URL(fn, contentBase)
        a.innerText = fn
        li.appendChild(a)
        document.getElementById("files").appendChild(li)
    }

    let baseElement = document.head.appendChild(document.createElement("base"))
    baseElement.href = contentBase

    window.app.puzzle = puzzle
    console.info("window.app.puzzle =", window.app.puzzle)

    console.groupEnd()
}

function init() {
    window.app = {}
    window.setanswer = setanswer

    for (let form of document.querySelectorAll("form.answer")) {
        form.addEventListener("submit", formSubmitHandler)
        for (let e of form.querySelectorAll("[name=answer]")) {
            e.addEventListener("input", answerInputHandler)
        }
    }
    // There isn't a more graceful way to "unload" scripts attached to the current puzzle
    window.addEventListener("hashchange", () => location.reload())

    // Make all links absolute, because we're going to be changing the base URL
    for (let e of document.querySelectorAll("[href]")) {
        e.href = new URL(e.href, location)
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

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
