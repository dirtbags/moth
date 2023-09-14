/**
 * Functionality for index.html (Login / Puzzles list)
 */
import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

class App {
    constructor(basePath=".") {
        this.server = new moth.Server(basePath)

        let uuid = Math.floor(Math.random() * 1000000).toString(16)
        this.fakeRegistration = {
            TeamId: uuid,
            TeamName: `Team ${uuid}`,
        }

        for (let form of document.querySelectorAll("form.login")) {
            form.addEventListener("submit", event => this.handleLoginSubmit(event))
        }
        for (let e of document.querySelectorAll(".logout")) {
            e.addEventListener("click", () => this.Logout())
        }

        setInterval(() => this.Update(), common.Minute/3)
        this.Update()
    }

    handleLoginSubmit(event) {
        event.preventDefault()
        console.log(event)
    }
    
    /**
     * Attempt to log in to the server.
     * 
     * @param {String} teamId 
     * @param {String} teamName 
     */
    async Login(teamId, teamName) {
        try {
            await this.server.Login(teamId, teamName)
            common.Toast(`Logged in (team id = ${teamId})`)
            this.Update()
        }
        catch (error) {
            common.Toast(error)
        }
    }

    /**
     * Log out of the server by clearing the saved Team ID.
     */
    async Logout() {
        try {
            this.server.Reset()
            common.Toast("Logged out")
            this.Update()
        }
        catch (error) {
            common.Toast(error)
        }
    }

    /**
     * Update the entire page.
     *
     * Fetch a new state, and rebuild all dynamic elements on this bage based on
     * what's returned. If we're in development mode and not logged in, auto
     * login too.
     */
    async Update() {
        this.state = await this.server.GetState()
        for (let e of document.querySelectorAll(".messages")) {
            e.innerHTML = this.state.Messages
        }

        for (let e of document.querySelectorAll(".login")) {
            this.renderLogin(e, !this.server.LoggedIn())
        }
        for (let e of document.querySelectorAll(".puzzles")) {
            this.renderPuzzles(e, this.server.LoggedIn())
        }

        if (this.state.DevelopmentMode() && !this.server.LoggedIn()) {
            common.Toast("Automatically logging in to devel server")
            console.info("Logging in with generated Team ID and Team Name", this.fakeRegistration)
            return this.Login(this.fakeRegistration.TeamId, this.fakeRegistration.TeamName)
        }
    }

    /**
     * Render a login box.
     * 
     * This just toggles visibility, there's nothing dynamic in a login box.
     */
    renderLogin(element, visible) {
        element.classList.toggle("hidden", !visible)
    }

    /**
     * Render a puzzles box.
     *
     * This updates the list of open puzzles, and adds mothball download links
     * if the server is in development mode.
     */
    renderPuzzles(element, visible) {
        element.classList.toggle("hidden", !visible)
        while (element.firstChild) element.firstChild.remove()
        for (let cat of this.state.Categories()) {
            let pdiv = element.appendChild(document.createElement("div"))
            pdiv.classList.add("category")
            
            let h = pdiv.appendChild(document.createElement("h2"))
            h.textContent = cat
            
            // Extras if we're running a devel server
            if (this.state.DevelopmentMode()) {
                let a = h.appendChild(document.createElement('a'))
                a.classList.add("mothball")
                a.textContent = "📦"
                a.href = this.server.URL(`mothballer/${cat}.mb`)
                a.title = "Download a compiled puzzle for this category"
            }
            
            // List out puzzles in this category
            let l = pdiv.appendChild(document.createElement("ul"))
            for (let puzzle of this.state.Puzzles(cat)) {
                let i = l.appendChild(document.createElement("li"))

                let url = new URL("puzzle.html", window.location)
                url.hash = `${puzzle.Category}:${puzzle.Points}`
                let a = i.appendChild(document.createElement("a"))
                a.textContent = puzzle.Points
                a.href = url
                a.target = "_blank"
            }

            if (!this.state.HasUnsolved(cat)) {
                l.appendChild(document.createElement("li")).textContent = "✿"
            }
            
            element.appendChild(pdiv)        
        }
    }
}

function init() {
    window.app = {
        server: new App()
    }
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
  