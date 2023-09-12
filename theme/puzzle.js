// jshint asi:true

// prettify adds classes to various types, returning an HTML string.
function prettify(key, val) {
  switch (key) {
    case "Body":
      return '[HTML]'
  }
  return val
}

// devel_addin drops a bunch of development extensions into element e.
// It will only modify stuff inside e.
function devel_addin(e) {
  let h = e.appendChild(document.createElement("h2"))
  h.textContent = "Developer Output"

  let log = window.puzzle.Debug.Log || []
  if (log.length > 0) {
    e.appendChild(document.createElement("h3")).textContent = "Log"
    let le = e.appendChild(document.createElement("ul"))
    for (let entry of log) {
      le.appendChild(document.createElement("li")).textContent = entry
    }
  }

  e.appendChild(document.createElement("h3")).textContent = "Puzzle object"
  
  let hobj = JSON.stringify(window.puzzle, prettify, 2)
  let d = e.appendChild(document.createElement("pre"))
  d.classList.add("object")
  d.innerHTML = hobj

  e.appendChild(document.createElement("p")).textContent = "This debugging information will not be available to participants."
}

// Hash routine used in v3.4 and earlier
function djb2hash(buf) {
  let h = 5381
  for (let c of (new TextEncoder()).encode(buf)) { // Encode as UTF-8 and read in each byte
    // JavaScript converts everything to a signed 32-bit integer when you do bitwise operations.
    // So we have to do "unsigned right shift" by zero to get it back to unsigned.
    h = (((h * 33) + c) & 0xffffffff) >>> 0
  }
  return h
}

// The routine used to hash answers in compiled puzzle packages
async function sha256Hash(message) {
  const msgUint8 = new TextEncoder().encode(message);                           // encode as (utf-8) Uint8Array
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgUint8);           // hash the message
  const hashArray = Array.from(new Uint8Array(hashBuffer));                     // convert buffer to byte array
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join(''); // convert bytes to hex string
  return hashHex;
}

// Is the provided answer possibly correct?
async function checkAnswer(answer) {
  let answerHashes = []
  answerHashes.push(djb2hash(answer))
  answerHashes.push(await sha256Hash(answer))

  for (let hash of answerHashes) {
    for (let correctHash of window.puzzle.AnswerHashes) {    
      if (hash == correctHash) {
        return true
      }
    }
  }
  return false
}

// Pop up a message
function toast(message, timeout=5000) {
  let p = document.createElement("p")
  
  p.innerText = message
  document.getElementById("messages").appendChild(p)
  setTimeout(
    e => { p.remove() },
    timeout
  )
}

// When the user submits an answer
function submit(e) {
  e.preventDefault()
  let data = new FormData(e.target)
  
  window.data = data
  fetch("answer", {
    method: "POST",
    body: data,
  })
  .then(resp => {
    if (resp.ok) {
      resp.json()
      .then(obj => {
        toast(obj.data.description)
      })
    } else {
      toast("Error submitting your answer. Try again in a few seconds.")
      console.log(resp)
    }
  })
  .catch(err => {
    toast("Error submitting your answer. Try again in a few seconds.")
    console.log(err)
  })
}

async function loadPuzzle(categoryName, points, puzzleId) {
  let puzzle = document.getElementById("puzzle")
  let base = "content/" + categoryName + "/" + puzzleId + "/"

  let resp = await fetch(base + "puzzle.json")
  if (! resp.ok) {
    console.log(resp)
    let err = await resp.text()
    Array.from(puzzle.childNodes).map(e => e.remove())
    p = puzzle.appendChild(document.createElement("p"))
    p.classList.add("Error")
    p.textContent = err
    return
  }

  // Make the whole puzzle available
  window.puzzle = await resp.json()
  
  // Populate authors
  document.getElementById("authors").textContent = window.puzzle.Authors.join(", ")

  // If answers are provided, this is the devel server
  if (window.puzzle.Answers.length > 0) {
    devel_addin(document.getElementById("devel"))
  }
  
  // Load scripts
  for (let script of (window.puzzle.Scripts || [])) {
    let st = document.createElement("script")
    document.head.appendChild(st)
    st.src = base + script
  }
  
  // List associated files
  for (let fn of (window.puzzle.Attachments || [])) {
    let li = document.createElement("li")
    let a = document.createElement("a")
    a.href = base + fn
    a.innerText = fn
    li.appendChild(a)
    document.getElementById("files").appendChild(li)
  }

  // Prefix `base` to relative URLs in the puzzle body
  let doc = new DOMParser().parseFromString(window.puzzle.Body, "text/html")
  for (let se of doc.querySelectorAll("[src],[href]")) {
    se.outerHTML = se.outerHTML.replace(/(src|href)="([^/]+)"/i, "$1=\"" + base + "$2\"")
  }
  
  // If a validation pattern was provided, set that
  if (window.puzzle.AnswerPattern) {
    document.querySelector("#answer").pattern = window.puzzle.AnswerPattern
  }

  // Replace puzzle children with what's in `doc`
  Array.from(puzzle.childNodes).map(e => e.remove())
  Array.from(doc.body.childNodes).map(e => puzzle.appendChild(e))
  
  document.title = categoryName + " " + points
  document.querySelector("body > h1").innerText = document.title
  document.querySelector("input[name=cat]").value = categoryName
  document.querySelector("input[name=points]").value = points
}

// Check to see if the answer might be correct
// This might be better done with the "constraint validation API"
// https://developer.mozilla.org/en-US/docs/Learn/HTML/Forms/Form_validation#Validating_forms_using_JavaScript
function answerCheck(e) {
  let answer = e.target.value
  let ok = document.querySelector("#answer_ok")
  
  // You have to provide someplace to put the check
  if (! ok) {
    return
  }
  
  checkAnswer(answer)
  .then (correct => {
    if (correct) {
      ok.textContent = "⭕"
      ok.title = "Possibly correct"
    } else {
      ok.textContent = "❌"
      ok.title = "Definitely not correct"
    }
  })
}

function init() {
  let params = new URLSearchParams(window.location.search)
  let categoryName = params.get("cat")
  let points = params.get("points")
  let puzzleId = params.get("pid")
  
  if (categoryName && points) {
    loadPuzzle(categoryName, points, puzzleId || points)
  }

  let teamId = sessionStorage.getItem("id")
  if (teamId) {
    document.querySelector("input[name=id]").value = teamId
  }
  
  if (document.querySelector("#answer")) {
    document.querySelector("#answer").addEventListener("input", answerCheck)
  }
  document.querySelector("form").addEventListener("submit", submit)
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init)
} else {
  init()
}
