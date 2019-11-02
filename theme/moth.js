// jshint asi:true

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

function renderPuzzles(obj) {
  let puzzlesElement = document.createElement('div')
  
  // Create a sorted list of category names
  let cats = Object.keys(obj)
  cats.sort()
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
    if (obj.__devel__) {
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
      let points = puzzle[0]
      let id = puzzle[1]
  
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
        a.href = "puzzle.html?cat=" + cat + "&points=" + points + "&pid=" + id
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

function heartbeat(teamId, participantId) {
  let url = new URL("puzzles.json", window.location)
  url.searchParams.set("id", teamId)
  if (participantId) {
    url.searchParams.set("pid", participantId)
  }
  let fd = new FormData()
  fd.append("id", teamId)
  fetch(url)
  .then(resp => {
    if (resp.ok) {
      resp.json()
      .then(renderPuzzles)
      .catch(err => {
        toast("Error fetching recent puzzles. I'll try again in a moment.")
        console.log(err)
      })
    }
  })
  .catch(err => {
    toast("Error fetching recent puzzles. I'll try again in a moment.")
    console.log(err)
  })
}

function showPuzzles(teamId, participantId) {
  let spinner = document.createElement("span")
  spinner.classList.add("spinner")

  sessionStorage.setItem("id", teamId)
  if (participantId) {
    sessionStorage.setItem("pid", participantId)
  }

  document.getElementById("login").style.display = "none"
  document.getElementById("puzzles").appendChild(spinner)
  heartbeat(teamId, participantId)
  setInterval(e => { heartbeat(teamId) }, 40000)
}

function login(e) {
  e.preventDefault()
  let name = document.querySelector("[name=name]").value
  let teamId = document.querySelector("[name=id]").value
  let pide = document.querySelector("[name=pid]")
  let participantId = pide?pide.value:""
  
  fetch("register", {
    method: "POST",
    body: new FormData(e.target),
  })
  .then(resp => {
    if (resp.ok) {
      resp.json()
      .then(obj => {
        if (obj.status == "success") {
          toast("Team registered")
          showPuzzles(teamId, participantId)
        } else if (obj.data.short == "Already registered") {
          toast("Logged in with previously-registered team name")
          showPuzzles(teamId, participantId)
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
  // Already signed in?
  let teamId = sessionStorage.getItem("id")
  let participantId = sessionStorage.getItem("pid")
  if (teamId) {
    showPuzzles(teamId, participantId)
  }
  
  document.getElementById("login").addEventListener("submit", login)
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init);
} else {
  init();
}
