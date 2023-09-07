/**
 * A point award.
 */
class Award {
    constructor(when, teamid, category, points) {
        /** Unix epoch timestamp for this award 
         * @type {Number}
        */
        this.When = when
        /** Team ID this award belongs to 
         * @type {String}
        */
        this.TeamID = teamid
        /** Puzzle category for this award
         * @type {String}
         */
        this.Category = category
        /** Points value of this award
         * @type {Number}
         */
        this.Points = points
    }
}

/**
 * A puzzle.
 * 
 * A new Puzzle only knows its category and point value.
 * If you want to populate it with meta-information, you must call Get().
 */
class Puzzle {
    /**
     * 
     * @param {Server} server 
     * @param {String} category 
     * @param {Number} points 
     */
    constructor (server, category, points) {
        if (points < 1) {
            throw(`Invalid points value: ${points}`)
        }
        
        /** Server where this puzzle lives
         * @type {Server}
         */
        this.server = server
        /** Category this puzzle belongs to
         * @type {String}
         */
        this.Category = category
        /** Point value of this puzzle
         * @type {Number}
         */
        this.Points = points
    }

    /** Error returned trying to fetch this puzzle */
    Error = {
        /** Status code provided by server */
        Status: 0,
        /** Status text provided by server */
        StatusText: "",
        /** Full text of server error */
        Body: "",
    }
    /** Hashes of answers 
     * @type {String[]}
     */
    AnswerHashes = []
    /** Pattern that answer should match
     * @type {String[]}
     */
   AnswerPattern = ""
    /** Accepted answers 
     * @type {String[]}
     */
    Answers = []
    /** Other files attached to this puzzles 
     * @type {String[]}
     */
    Attachments = []
    /** This puzzle's authors 
     * @type {String[]}
    */
    Authors = []
    /** HTML body of this puzzle */
    Body = ""
    /** Debugging information */
    Debug = {
        Errors: [],
        Hints: [],
        Log: [],
        Notes: "",
        Summary: "",
    }
    /** KSAs met by solving this puzzle 
     * @type {String[]}
    */
    KSAs = []
    /** Learning objective for this puzzle */
    Objective = "" 
    /** ECMAScript scripts needed for this puzzle 
     * @type {String[]}
    */
    Scripts = []
    /** Criteria for succeeding at this puzzle */
    Success = {
        /** Acceptable Minimum criteria for success */
        Minimum: "",
        /** Criteria for demonstrating mastery of this puzzle */
        Mastery: "",
    }

    /**
     * Populate this Puzzle object with meta-information from the server.
     */
    async Populate() {
        let resp = await this.Get("puzzle.json")
        if (!resp.ok) {
            let body = await resp.text()
            this.Error = {
                Status: resp.status,
                StatusText: resp.statusText,
                Body: body,
            }
            throw(this.Error)
        }
        let obj = await resp.json()
        Object.assign(this, obj)

        // Make sure lists are lists
        this.AnswerHashes ||= []
        this.Answers ||= []
        this.Attachments ||= []
        this.Authors ||= []
        this.Debug.Errors ||= []
        this.Debug.Hints ||= []
        this.Debug.Log ||= []
        this.KSAs ||= []
        this.Scripts ||= []
    }

    /**
     * Get a resource associated with this puzzle.
     * 
     * @param {String} filename Attachment/Script to retrieve
     * @returns {Promise<Response>}
     */
    Get(filename) {
        return this.server.GetContent(this.Category, this.Points, filename)
    }
}

/**
 * MOTH instance state.
 * 
 * @property {Object} Config
 * @property {Boolean} Config.Enabled Are points log updates enabled?
 * @property {String} Messages Global broadcast messages, in HTML
 * @property {Object.<String>} TeamNames Mapping from IDs to team names
 * @property {Object.<String,Number[]>} PointsByCategory Map from category name to open puzzle point values
 * @property {Award[]} PointsLog Log of points awarded
 */
class State {
    /**
     * @param {Server} server Server where we got this
     * @param {Object} obj Raw state data
     */
    constructor(server, obj) {
        for (let key of ["Config", "Messages", "TeamNames", "PointsLog"]) {
            if (!obj[key]) {
                throw(`Missing state property: ${key}`)
            }
        }
        this.server = server

        /** Configuration */
        this.Config = {
            /** Is the server in debug mode?
             * @type {Boolean}
             */
            Debug: obj.Config.Debug,
        }
        /** Global messages, in HTML
         * @type {String}
         */
        this.Messages = obj.Messages
        /** Map from Team ID to Team Name
         * @type {Object.<String,String>}
         */
        this.TeamNames = obj.TeamNames
        /** Map from category name to puzzle point values
         * @type {Object.<String,Number}
         */
        this.PointsByCategory = obj.Puzzles
        /** Log of points awarded
         * @type {Award[]}
         */
        this.PointsLog = obj.PointsLog.map((t,i,c,p) => new Award(t,i,c,p))
    }

    /**
     * Returns a sorted list of open category names
     * 
     * @returns {String[]} List of categories
     */
    Categories() {
        let ret = []
        for (let category in this.PointsByCategory) {
            ret.push(category)
        }
        ret.sort()
        return ret
    }

    /**
     * Check whether a category has unsolved puzzles.
     * 
     * The server adds a puzzle with 0 points in every "solved" category,
     * so this just checks whether there is a 0-point puzzle in the category's point list.
     * 
     * @param {String} category 
     * @returns {Boolean}
     */
    HasUnsolved(category) {
        return !this.PointsByCategory[category].includes(0)
    }

    /**
     * Return all open puzzles.
     * 
     * The returned list will be sorted by (category, points).
     * If not categories are given, all puzzles will be returned.
     * 
     * @param {String} categories Limit results to these categories
     * @returns {Puzzle[]}
     */
    Puzzles(...categories) {
        if (categories.length == 0) {
            categories = this.Categories()
        }
        let ret = []
        for (let category of categories) {
            for (let points of this.PointsByCategory[category]) {
                if (0 == points) {
                    // This means all potential puzzles in the category are open
                    continue
                }
                let p = new Puzzle(this.server, category, points)
                ret.push(p)
            }
        }
        return ret
    }
}

/**
 * A MOTH Server interface.
 * 
 * This uses localStorage to remember Team ID,
 * and will send a Team ID with every request, if it can find one.
 */
class Server {
    constructor(baseUrl) {
        this.baseUrl = new URL(baseUrl, location)
        this.teameIdKey = this.baseUrl.toString() + " teamID"
        this.teamId = localStorage[this.teameIdKey]
    }

    /**
     * Fetch a MOTH resource.
     * 
     * If anything other than a 2xx code is returned,
     * this function throws an error.
     * 
     * This always sends teamId.
     * If body is set, POST will be used instead of GET
     * 
     * @param {String} path Path to API endpoint
     * @param {Object<String,String>} body Key/Values to send in POST data
     * @returns {Promise<Response>} Response
     */
    fetch(path, body) {
        let url = new URL(path, this.baseUrl)
        if (this.teamId & (!(body && body.id))) {
            url.searchParams.set("id", this.teamId)
        }
        return fetch(url, {
            method: body?"POST":"GET",
            body,
        })
    }

    /**
     * Send a request to a JSend API endpoint.
     * 
     * @param {String} path Path to API endpoint
     * @param {Object<String,String>} args Key/Values to send in POST
     * @returns {Promise<Object>} JSend Data
     */
    async call(path, args) {
        let resp = await this.fetch(path, args)
        let obj = await resp.json()
        switch (obj.status) {
            case "success":
                return obj.data
            case "failure":
                throw new Error(obj.data.description || obj.data.short || obj.data)
            case "error":
                throw new Error(obj.message)
            default:
                throw new Error(`Unknown JSend status: ${obj.status}`)
        }
    }

    /**
     * Forget about any previous Team ID.
     * 
     * This is equivalent to logging out.
     */
    Reset() {
        localStorage.removeItem(this.teameIdKey)
        this.teamId = null
    }

    /**
     * Fetch current contest state.
     * 
     * @returns {State}
     */
    async GetState() {
        let resp = await this.fetch("/state")
        let obj = await resp.json()
        return new State(this, obj)
    }

    /**
     * Register a team name with a team ID.
     * 
     * This is similar to, but not exactly the same as, logging in.
     * See MOTH documentation for details.
     * 
     * @param {String} teamId 
     * @param {String} teamName 
     * @returns {Promise<String>} Success message from server
     */
    async Register(teamId, teamName) {
        let data = await this.call("/login", {id: teamId, name: teamName})
        this.teamId = teamId
        this.teamName = teamName
        localStorage[this.teameIdKey] = teamId
        return data.description || data.short
    }

    /**
     * Submit a puzzle answer for points.
     *
     * The returned promise will fail if anything goes wrong, including the
     * answer being rejected.
     *
     * @param {String} category Category of puzzle
     * @param {Number} points Point value of puzzle
     * @param {String} answer Answer to submit
     * @returns {Promise<Boolean>} Was the answer accepted?
     */
    async SubmitAnswer(category, points, answer) {
        await this.call("/answer", {category, points, answer})
        return true
    }

    /**
     * Fetch a file associated with a puzzle.
     * 
     * @param {String} category Category of puzzle
     * @param {Number} points Point value of puzzle
     * @param {String} filename
     * @returns {Promise<Response>}
     */
    GetContent(category, points, filename) {
        return this.fetch(`/content/${category}/${points}/${filename}`)
    }
}

export {
    Server
}