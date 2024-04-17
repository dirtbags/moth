/**
 * Functionality for index.html (Login / Puzzles list)
 */
import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

class App {
    constructor(basePath=".") {        
        this.config = {}
        
        this.server = new moth.Server(basePath)

        for (let form of document.querySelectorAll("form.login")) {
            form.addEventListener("submit", event => this.handleLoginSubmit(event))
        }
        for (let e of document.querySelectorAll(".logout")) {
            e.addEventListener("click", () => this.Logout())
        }

        common.StateUpdateChannel.addEventListener("message", () => {
            // Give mothd time to catch up
            setTimeout(() => this.UpdateState(), 1/2 * common.Second)
        })

        setInterval(() => this.UpdateState(), common.Minute/3)
        setInterval(() => this.UpdateConfig(), common.Minute* 5)

        this.UpdateConfig()
        .finally(() => this.UpdateState())
    }

    handleLoginSubmit(event) {
        event.preventDefault()
        let f = new FormData(event.target)
        this.Login(f.get("id"), f.get("name"))
    }
    
    /**
     * Attempt to log in to the server.
     * 
     * @param {string} teamID 
     * @param {string} teamName 
     */
    async Login(teamID, teamName) {
        try {
            await this.server.Login(teamID, teamName)
            common.Toast(`Logged in (team id = ${teamID})`)
            this.UpdateState()
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
            this.UpdateState()
        }
        catch (error) {
            common.Toast(error)
        }
    }

    /**
     * Update app configuration.
     *
     * Configuration can be updated less frequently than state, to reduce server
     * load, since configuration should (hopefully) change less frequently.
     */
    async UpdateConfig() {
        this.config = await common.Config()

        for (let e of document.querySelectorAll(".messages")) {
            e.innerHTML = this.config.Messages || ""
        }
    }

    /**
     * Update the entire page.
     *
     * Fetch a new state, and rebuild all dynamic elements on this bage based on
     * what's returned. If we're in development mode and not logged in, auto
     * login too.
     */
    async UpdateState() {
        this.state = await this.server.GetState()

        // Update elements with data-track-solved
        for (let e of document.querySelectorAll("[data-track-solved]")) {
            // Only hide if data-track-solved is different than config.PuzzleList.TrackSolved
            let tracking = this.config.PuzzleList?.TrackSolved || false
            let displayIf = common.StringTruthy(e.dataset.trackSolved)
            e.classList.toggle("hidden", tracking != displayIf)
        }

        for (let e of document.querySelectorAll(".login")) {
            this.renderLogin(e, !this.server.LoggedIn())
        }
        for (let e of document.querySelectorAll(".puzzles")) {
            this.renderPuzzles(e, this.server.LoggedIn())
        }

        if (this.state.DevelopmentMode() && !this.server.LoggedIn()) {
            let teamID = Math.floor(Math.random() * 1000000).toString(16)
            common.Toast("Automatically logging in to devel server")
            console.info(`Logging in with generated Team ID: ${teamID}`)
            return this.Login(teamID, `Team ${teamID}`)
        }
    }

    /**
     * Render a login box.
     * 
     * Just toggles visibility, there's nothing dynamic in a login box.
     */
    renderLogin(element, visible) {
        element.classList.toggle("hidden", !visible)
    }

    /**
     * Render a puzzles box.
     *
     * Displays the list of open puzzles, and adds mothball download links
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
                a.textContent = "⬇️"
                a.href = this.server.URL(`mothballer/${cat}.mb`)
                a.title = "Download a compiled puzzle for this category"
            }
            
            // List out puzzles in this category
            let l = pdiv.appendChild(document.createElement("ul"))
            for (let puzzle of this.state.Puzzles(cat)) {
                let i = l.appendChild(document.createElement("li"))

                let url = new URL("puzzle.html", common.BaseURL)
                url.hash = `${puzzle.Category}:${puzzle.Points}`
                let a = i.appendChild(document.createElement("a"))
                a.textContent = puzzle.Points
                a.href = url
                a.target = "_blank"

                if (this.config.PuzzleList?.TrackSolved) {
                    a.classList.toggle("solved", this.state.IsSolved(puzzle))
                }
                if (this.config.PuzzleList?.Titles) {
                    this.loadTitle(puzzle, i)
                }
            }

            if (!this.state.ContainsUnsolved(cat)) {
                l.appendChild(document.createElement("li")).textContent = "✿"
            }
            
            element.appendChild(pdiv)        
        }
    }

    /**
     * Asynchronously loads a puzzle, in order to populate the title.
     * 
     * Calling this for every open puzzle will generate a lot of load on the server.
     * If we decide we want this for a multi-participant server,
     * we should implement some sort of cache.
     * 
     * @param {Puzzle} puzzle 
     * @param {Element} element 
     */
    async loadTitle(puzzle, element) {
        await puzzle.Populate()
        let title = puzzle.Extra.title
        if (!title) {
            return
        }
        element.classList.add("entitled")
        for (let a of element.querySelectorAll("a")) {
            a.textContent += `: ${title}`
        }
    }
}

function init() {
    window.app = new App()
}

common.WhenDOMLoaded(init)
