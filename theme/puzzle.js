// jshint asi:true

// devel_addin drops a bunch of development extensions into element e.
// It will only modify stuff inside e.
function devel_addin(obj, e) {
  let h = document.createElement("h2")
  e.appendChild(h)
  h.textContent = "Development Options"

  let g = document.createElement("p")
  e.appendChild(g)
  g.innerText = "This section will not appear for participants."
  
  let keys = Object.keys(obj)
  keys.sort()
  for (let key of keys) {
    switch (key) {
      case "body":
        continue
    }
    let val = obj[key]

    if ((! val) || (val.length === 0)) {
      // Empty, skip it
      continue
    }

    let d = document.createElement("div")
    e.appendChild(d)
    d.classList.add("kvpair")
    
    let ktxt = document.createElement("span")
    d.appendChild(ktxt)
    ktxt.textContent = key
    
    if (Array.isArray(val)) {
      let vi = document.createElement("select")
      d.appendChild(vi)
      vi.multiple = true
      for (let a of val) {
        let opt = document.createElement("option")
        vi.appendChild(opt)
        opt.innerText = a
      }
    } else {
      let vi = document.createElement("input")
      d.appendChild(vi)
      vi.value = val
      vi.disabled = true
    }
  }
}



// The routine used to hash answers in compiled puzzle packages
function djb2hash(buf) {
  let h = 5381
  for (let c of (new TextEncoder).encode(buf)) { // Encode as UTF-8 and read in each byte
    // JavaScript converts everything to a signed 32-bit integer when you do bitwise operations.
    // So we have to do "unsigned right shift" by zero to get it back to unsigned.
    h = (((h * 33) + c) & 0xffffffff) >>> 0
  }
  return h
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
  fetch("answer", {
    method: "POST",
    body: new FormData(e.target),
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

function loadPuzzle(categoryName, points, puzzleId) {
  let puzzle = document.getElementById("puzzle")
  let base = "content/" + categoryName + "/" + puzzleId + "/"

  fetch(base + "puzzle.json")
  .then(resp => {
    return resp.json()
  })
  .then(obj => {
    // Populate authors
    document.getElementById("authors").textContent = obj.authors.join(", ")
    
    // Make the whole puzzle available
    window.puzzle = obj
    
    // If answers are provided, this is the devel server
    if (obj.answers) {
      devel_addin(obj, document.getElementById("devel"))
    }
    
    // Load scripts
    for (let script of obj.scripts) {
      let st = document.createElement("script")
      document.head.appendChild(st)
      st.src = base + script
    }
    
    // List associated files
    for (let fn of obj.files) {
      let li = document.createElement("li")
      let a = document.createElement("a")
      a.href = base + fn
      a.innerText = fn
      li.appendChild(a)
      document.getElementById("files").appendChild(li)
    }

    // Prefix `base` to relative URLs in the puzzle body
    let doc = new DOMParser().parseFromString(obj.body, "text/html")
    for (let se of doc.querySelectorAll("[src],[href]")) {
      se.outerHTML = se.outerHTML.replace(/(src|href)="([^/]+)"/i, "$1=\"" + base + "$2\"")
    }
    
    // If a validation pattern was provided, set that
    if (obj.pattern) {
      document.querySelector("#answer").pattern = obj.pattern
    }

    // Replace puzzle children with what's in `doc`
    Array.from(puzzle.childNodes).map(e => e.remove())
    Array.from(doc.body.childNodes).map(e => puzzle.appendChild(e))
  })
  .catch(err => {
    // Show error to the user
    Array.from(puzzle.childNodes).map(e => e.remove())
    let p = document.createElement("p")
    puzzle.appendChild(p)
    p.classList.add("Error")
    p.textContent = err
  })
  
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
  
  let possiblyCorrect = false
  let answerHash = djb2hash(answer)
  for (let correctHash of window.puzzle.hashes) {
    if (correctHash == answerHash) {
      possiblyCorrect = true
    }
  }
  
  if (possiblyCorrect) {
    ok.textContent = "❓"
    ok.title = "Possibly correct"
  } else {
    ok.textContent = "⛔"
    ok.title = "Definitely not correct"
  }
}

function init() {
  let params = new URLSearchParams(window.location.search)
  let categoryName = params.get("cat")
  let points = params.get("points")
  let puzzleId = params.get("pid")
  
  if (categoryName && points && puzzleId) {
    loadPuzzle(categoryName, points, puzzleId)
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

