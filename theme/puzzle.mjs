/**
 * Functionality for puzzle.html (Puzzle display / answer form)
 */
import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

const server = new moth.Server(".")

/**
 * Handle a submit event on a form.
 * 
 * Called when the user submits the form,
 * either by clicking a "submit" button,
 * or by some other means provided by the browser,
 * like hitting the Enter key.
 * 
 * @param {Event} event 
 */
async function formSubmitHandler(event) {
    event.preventDefault()
    let data = new FormData(event.target)
    let proposed = data.get("answer")
    let message
    
    console.groupCollapsed("Submit answer")
    console.info(`Proposed answer: ${proposed}`)
    try {
        message = await window.app.puzzle.SubmitAnswer(proposed)
        common.Toast(message)
    }
    catch (err) {
        common.Toast(err)
    }
    console.groupEnd("Submit answer")
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
 * @param {boolean} clear Should the element be cleared of children? Default true.
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
 * Errors are rendered in the puzzle area, so the user can see a bit more about
 * what the problem is.
 *
 * @param {string} error 
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
 *  Makes sure the Circle Of Success gets updated.
 * 
 * @param {string} s 
 */
function SetAnswer(s) {
    let e = document.querySelector("#answer")
    e.value = s
    e.dispatchEvent(new Event("input"))
}

function writeObject(e, obj) {
    let keys = Object.keys(obj)
    keys.sort()
    for (let key of keys) {
        let val = obj[key]
        if ((key === "Body") || (!val) || (val.length === 0)) {
            continue
        }

        let d = e.appendChild(document.createElement("dt"))
        d.textContent = key

        let t = e.appendChild(document.createElement("dd"))
        if (Array.isArray(val)) {
            let vi = t.appendChild(document.createElement("ul"))
            vi.multiple = true
            for (let a of val) {
                let opt = vi.appendChild(document.createElement("li"))
                opt.textContent = a
            }
        } else if (typeof(val) === "object") {
            writeObject(t, val)
        } else {
            t.textContent = val
        }
    }
}

/**
 * Load the given puzzle.
 * 
 * @param {string} category 
 * @param {number} points 
 */
async function loadPuzzle(category, points) {
    console.groupCollapsed("Loading puzzle:", category, points)
    let contentBase = new URL(`content/${category}/${points}/`, common.BaseURL)
    
    // Tell user we're loading
    puzzleElement().appendChild(document.createElement("progress"))
    for (let qs of ["#authors", "#title", "title"]) {
        for (let e of document.querySelectorAll(qs)) {
            e.textContent = "[loading]"
        }
    }    

    let puzzle = server.GetPuzzle(category, points)

    console.time("Populate")
    try {
        await puzzle.Populate()
    }
    catch {
        let error = puzzleElement().appendChild(document.createElement("pre"))
        error.classList.add("notification", "error")
        error.textContent = puzzle.Error.Body
        return
    }
    finally {
        console.timeEnd("Populate")
    }

    console.info(`Setting base tag to ${contentBase}`)
    let baseElement = document.head.appendChild(document.createElement("base"))
    baseElement.href = contentBase

    console.info("Tweaking HTML...")
    let title = `${category} ${points}`
    document.querySelector("title").textContent = title
    document.querySelector("#title").textContent = title
    document.querySelector("#authors").textContent = puzzle.Authors.join(", ")
    if (puzzle.AnswerPattern) {
        document.querySelector("#answer").pattern = puzzle.AnswerPattern
    }
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


    console.info("Filling debug information...")
    for (let e of document.querySelectorAll(".debug")) {
        if (puzzle.Answers.length > 0) {
            writeObject(e, puzzle)
        } else {
            e.classList.add("hidden")
        }
    }

    window.app.puzzle = puzzle
    console.info("window.app.puzzle =", window.app.puzzle)

    console.groupEnd()

    return puzzle
}

async function init() {
    window.app = {}
    window.setanswer = (str => SetAnswer(str))

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
        e.href = new URL(e.href, common.BaseURL)
    }

    let hashpart = location.hash.split("#")[1] || ""
    let catpoints = hashpart.split(":")
    let category = catpoints[0]
    let points = Number(catpoints[1])
    if (!category && !points) {
        error(`Doesn't look like a puzzle reference: ${hashpart}`)
        return
    }

    window.app.puzzle = await loadPuzzle(category, points)
}

common.WhenDOMLoaded(init)
