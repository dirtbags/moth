// jshint asi:true

var devel = false
var teamId
var heartbeatInterval = 40000

function toast(message, timeout=5000) {
  let p = document.createElement("p")
  
  p.innerText = message
  document.getElementById("messages").appendChild(p)
  setTimeout(
    e => { p.remove() },
    timeout
  )
}

function renderNotices(obj) {
  let ne = document.getElementById("notices")
  if (ne) {
    ne.innerHTML = obj
  }
}

function renderPuzzles(obj) {
  let puzzlesElement = document.createElement('div')
  
  document.getElementById("login").style.display = "none"
  
  // Create a sorted list of category names
  let cats = Object.keys(obj)
  cats.sort()
  if (cats.length == 0) {
    toast("No categories to serve!")
  }
  for (let cat of cats) {
    if (cat.startsWith("__")) {
      // Skip metadata
      continue
    }
    let puzzles = obj[cat]
    
    let pdiv = document.createElement('div')
    pdiv.className = 'category'
    
    let h = document.createElement('h2')
    pdiv.appendChild(h)
    h.textContent = cat
    
    // Extras if we're running a devel server
    if (devel) {
      let a = document.createElement('a')
      h.insertBefore(a, h.firstChild)
      a.textContent = "⬇️"
      a.href = "mothballer/" + cat + ".mb"
      a.classList.add("mothball")
      a.title = "Download a compiled puzzle for this category"
    }
    
    // List out puzzles in this category
    let l = document.createElement('ul')
    pdiv.appendChild(l)
    for (let puzzle of puzzles) {
      let points = puzzle
      let id = null
      
      if (Array.isArray(puzzle)) {
        points = puzzle[0]
        id = puzzle[1]
      }
  
      let i = document.createElement('li')
      l.appendChild(i)
      i.textContent = " "
  
      if (points === 0) {
        // Sentry: there are no more puzzles in this category
        i.textContent = "✿"
      } else {
        let a = document.createElement('a')
        i.appendChild(a)
        a.textContent = points
        let url = new URL("puzzle.html", window.location)
        url.searchParams.set("cat", cat)
        url.searchParams.set("points", points)
        a.href = url.toString()
      }
    }
    
    puzzlesElement.appendChild(pdiv)
  }
    
  // Drop that thing in
  let container = document.getElementById("puzzles")
  while (container.firstChild) {
    container.firstChild.remove()
  }
  container.appendChild(puzzlesElement)
}

function renderState(obj) {
  window.state = obj
  devel = obj.Config.Devel
  if (devel) {
    let params = new URLSearchParams(window.location.search)
    sessionStorage.id = "1"
    renderPuzzles(obj.Puzzles)
  } else if (Object.keys(obj.Puzzles).length > 0) {
    renderPuzzles(obj.Puzzles)
  }
  renderNotices(obj.Messages)
}

function heartbeat() {
  let teamId = sessionStorage.id || ""
  let url = new URL("state", window.location)
  url.searchParams.set("id", teamId)
  let fd = new FormData()
  fd.append("id", teamId)
  fetch(url)
  .then(resp => {
    if (resp.ok) {
      resp.json()
      .then(renderState)
      .catch(err => {
        toast("Error fetching recent state. I'll try again in a moment.")
        console.log(err)
      })
    }
  })
  .catch(err => {
    toast("Error fetching recent state. I'll try again in a moment.")
    console.log(err)
  })
}

function showPuzzles() {
  let spinner = document.createElement("span")
  spinner.classList.add("spinner")

  document.getElementById("login").style.display = "none"
  document.getElementById("puzzles").appendChild(spinner)
}

function login(e) {
  e.preventDefault()
  let name = document.querySelector("[name=name]").value
  let teamId = document.querySelector("[name=id]").value
  
  fetch("register", {
    method: "POST",
    body: new FormData(e.target),
  })
  .then(resp => {
    if (resp.ok) {
      resp.json()
      .then(obj => {
        if ((obj.status == "success") || (obj.data.short == "Already registered")) {
          toast("Logged in")
          sessionStorage.id = teamId
          showPuzzles()
          heartbeat()
        } else {
          toast(obj.data.description)
        }
      })
      .catch(err => {
        toast("Oops, the server has lost its mind. You probably need to tell someone so they can fix it.")
        console.log(err, resp)
      })
    } else {
      toast("Oops, something's wrong with the server. Try again in a few seconds.")
      console.log(resp)
    }
  })
  .catch(err => {
    toast("Oops, something went wrong. Try again in a few seconds.")
    console.log(err)
  })
}

function init() {
  heartbeat()
  setInterval(e => heartbeat(), 40000)

  document.getElementById("login").addEventListener("submit", login)
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init)
} else {
  init()
}

