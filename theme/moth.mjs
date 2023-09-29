/**
 * Hash/digest functions
 */
class Hash {
    /**
     * Dan Bernstein hash
     * 
     * Used until MOTH v3.5
     * 
     * @param {string} buf Input
     * @returns {number}
     */
    static djb2(buf) {
        let h = 5381
        for (let c of (new TextEncoder()).encode(buf)) { // Encode as UTF-8 and read in each byte
            // JavaScript converts everything to a signed 32-bit integer when you do bitwise operations.
            // So we have to do "unsigned right shift" by zero to get it back to unsigned.
            h = ((h * 33) + c) >>> 0
        }
        return h
    }

    /**
     * Dan Bernstein hash with xor
     * 
     * @param {string} buf Input
     * @returns {number}
     */
    static djb2xor(buf) {
        let h = 5381
        for (let c of (new TextEncoder()).encode(buf)) {
            h = ((h * 33) ^ c) >>> 0
        }
        return h
    }
  
    /**
     * SHA 256
     * 
     * Used until MOTH v4.5
     * 
     * @param {string} buf Input
     * @returns {Promise.<string>} hex-encoded digest
     */
    static async sha256(buf) {
        const msgUint8 = new TextEncoder().encode(buf)
        const hashBuffer = await crypto.subtle.digest('SHA-256', msgUint8)
        const hashArray = Array.from(new Uint8Array(hashBuffer))
        return this.hexlify(hashArray);
    }

    /**
     * SHA 1, but only the first 4 hexits (2 octets).
     * 
     * Git uses this technique with 7 hexits (default) as a "short identifier".
     * 
     * @param {string} buf Input
     */
    static async sha1_slice(buf, end=4) {
        const msgUint8 = new TextEncoder().encode(buf)
        const hashBuffer = await crypto.subtle.digest("SHA-1", msgUint8)
        const hashArray = Array.from(new Uint8Array(hashBuffer))
        const hexits = this.hexlify(hashArray)
        return hexits.slice(0, end)
    }

    /**
     * Hex-encode a byte array
     * 
     * @param {number[]} buf Byte array
     * @returns {string}
     */
    static hexlify(buf) {
        return buf.map(b => b.toString(16).padStart(2, "0")).join("")   
    }

  /**
   * Apply every hash to the input buffer.
   * 
   * @param {string} buf Input
   * @returns {Promise.<string[]>}
   */
  static async All(buf) {
    return [
        String(this.djb2(buf)),
        await this.sha256(buf),
        await this.sha1_slice(buf),
    ]
  }
}

/**
 * A point award.
 */
class Award {
    constructor(when, teamid, category, points) {
        /** Unix epoch timestamp for this award 
         * @type {number}
        */
        this.When = when
        /** Team ID this award belongs to 
         * @type {string}
        */
        this.TeamID = teamid
        /** Puzzle category for this award
         * @type {string}
         */
        this.Category = category
        /** Points value of this award
         * @type {number}
         */
        this.Points = points
    }
}

/**
 * A puzzle.
 * 
 * A new Puzzle only knows its category and point value.
 * If you want to populate it with meta-information, you must call Populate().
 * 
 * Parameters created by Populate are described in the server source code:
 * {@link https://pkg.go.dev/github.com/dirtbags/moth/v4/pkg/transpile#Puzzle}
 * 
 */
class Puzzle {
    /**
     * @param {Server} server 
     * @param {string} category 
     * @param {number} points 
     */
    constructor (server, category, points) {
        if (points < 1) {
            throw(`Invalid points value: ${points}`)
        }
        
        /** Server where this puzzle lives
         * @type {Server}
         */
        this.server = server
        
        /** Category this puzzle belongs to */
        this.Category = String(category)
        
        /** Point value of this puzzle */
        this.Points = Number(points)

        /** Error returned trying to retrieve this puzzle */
        this.Error = {
            /** Status code provided by server */
            Status: 0,
            /** Status text provided by server */
            StatusText: "",
            /** Full text of server error */
            Body: "",
        }
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
     * @param {string} filename Attachment/Script to retrieve
     * @returns {Promise.<Response>}
     */
    Get(filename) {
        return this.server.GetContent(this.Category, this.Points, filename)
    }

    /**
     * Check if a string is possibly correct.
     *
     * The server sends a list of answer hashes with each puzzle: this method
     * checks to see if any of those hashes match a hash of the string.
     *
     * The MOTH development team likes obscure hash functions with a lot of
     * collisions, which means that a given input may match another possible
     * string's hash. We do this so that if you run a brute force attack against
     * the list of hashes, you have to write your own brute force program, and
     * you still have to pick through a lot of potentially correct answers when
     * it's done.
     *
     * @param {string} str User-submitted possible answer
     * @returns {Promise.<boolean>}
     */
    async IsPossiblyCorrect(str) {
        let userAnswerHashes = await Hash.All(str)

        for (let pah of this.AnswerHashes) {
            for (let uah of userAnswerHashes) {
                if (pah == uah) {
                    return true
                }
            }
        }
        return false
    }

    /**
     * Submit a proposed answer for points.
     *
     * The returned promise will fail if anything goes wrong, including the
     * proposed answer being rejected.
     *
     * @param {string} proposed Answer to submit
     * @returns {Promise.<string>} Success message
     */
    SubmitAnswer(proposed) {
        return this.server.SubmitAnswer(this.Category, this.Points, proposed)
    }
}

/**
 * A snapshot of scores.
 */
class Scores {
    constructor() {
        /** 
         * Timestamp of this score snapshot
         * @type number 
         */
        this.Timestamp = 0

        /**
         * All categories present in this snapshot.
         *
         * ECMAScript sets preserve order, so iterating over this will yield
         * categories as they were added to the points log.
         *
         * @type {Set.<string>}
         */
        this.Categories = new Set()

        /**
         * All team IDs present in this snapshot
         * @type {Set.<string>}
         */
        this.TeamIDs = new Set()

        /**
         * Highest score in each category
         * @type {Object.<string,number>}
         */
        this.MaxPoints = {}
        
        this.categoryTeamPoints = {}
    }

    /**
     * Return a sorted list of category names
     * 
     * @returns {string[]}
     */
    SortedCategories() {
        let categories = [...this.Categories]
        categories.sort((a,b) => a.localeCompare(b, "en", {sensitivity: "base"}))
        return categories
    }

    /**
     * Add an award to a team's score.
     * 
     * Updates this.Timestamp to the award's timestamp.
     * 
     * @param {Award} award 
     */
    Add(award) {
        this.Timestamp = award.Timestamp
        this.Categories.add(award.Category)
        this.TeamIDs.add(award.TeamID)

        let teamPoints = (this.categoryTeamPoints[award.Category] ??= {})
        let points = (teamPoints[award.TeamID] || 0) + award.Points
        teamPoints[award.TeamID] = points

        let max = this.MaxPoints[award.Category] || 0
        this.MaxPoints[award.Category] = Math.max(max, points)
    }

    /**
     * Get a team's score within a category.
     * 
     * @param {string} category 
     * @param {string} teamID 
     * @returns {number}
     */
    GetPoints(category, teamID) {
        let teamPoints = this.categoryTeamPoints[category] || {}
        return teamPoints[teamID] || 0
    }

    /**
     * Calculate a team's score in a category, using the Cyber Fire algorithm.
     * 
     *@param {string} category 
     * @param {string} teamID 
     */
    CyFiCategoryScore(category, teamID) {
        return this.GetPoints(category, teamID) / this.MaxPoints[category]
    }

    /**
     * Calculate a team's overall score, using the Cyber Fire algorithm.
     * 
     *@param {string} category 
     * @param {string} teamID 
     * @returns {number}
     */
    CyFiScore(teamID) {
        let score = 0
        for (let category of this.Categories) {
            score += this.CyFiCategoryScore(category, teamID)
        }
        return score
    }
}

/**
 * MOTH instance state.
 */
class State {
    /**
     * @param {Server} server Server where we got this
     * @param {Object} obj Raw state data
     */
    constructor(server, obj) {
        for (let key of ["Config", "TeamNames", "PointsLog"]) {
            if (!obj[key]) {
                throw(`Missing state property: ${key}`)
            }
        }
        this.server = server

        /** Configuration */
        this.Config = {
            /** Is the server in development mode?
             * @type {boolean}
             */
            Devel: obj.Config.Devel,
        }

        /** True if the server is in enabled state */
        this.Enabled = obj.Enabled

        /** Map from Team ID to Team Name
         * @type {Object.<string,string>}
         */
        this.TeamNames = obj.TeamNames

        /** Map from category name to puzzle point values
         * @type {Object.<string,number>}
         */
        this.PointsByCategory = obj.Puzzles

        /** Log of points awarded
         * @type {Award[]}
         */
        this.PointsLog = obj.PointsLog.map(entry => new Award(entry[0], entry[1], entry[2], entry[3]))
    }

    /**
     * Returns a sorted list of open category names
     * 
     * @returns {string[]} List of categories
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
     * Check whether a category contains unsolved puzzles.
     * 
     * The server adds a puzzle with 0 points in every "solved" category,
     * so this just checks whether there is a 0-point puzzle in the category's point list.
     * 
     * @param {string} category 
     * @returns {boolean}
     */
    ContainsUnsolved(category) {
        return !this.PointsByCategory[category].includes(0)
    }

    /**
     * Is the server in development mode?
     * 
     * @returns {boolean}
     */
    DevelopmentMode() {
        return this.Config && this.Config.Devel
    }

    /**
     * Return all open puzzles.
     * 
     * The returned list will be sorted by (category, points).
     * If not categories are given, all puzzles will be returned.
     * 
     * @param {string} categories Limit results to these categories
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

    /**
     * Has this puzzle been solved by this team?
     * 
     * @param {Puzzle} puzzle 
     * @param {string} teamID Team to check, default the logged-in team
     * @returns {boolean}
     */
    IsSolved(puzzle, teamID="self") {
        for (let award of this.PointsLog) {
            if (
                (award.Category == puzzle.Category)
                && (award.Points == puzzle.Points)
                && (award.TeamID == teamID)
            ) {
                return true
            }
        }
        return false
    }

    /**
     * Replay scores.
     *
     * MOTH has no notion of who is "winning", we consider this a user interface
     * decision. There are lots of interesting options: see
     * [scoring]{@link ../docs/scoring.md} for more.
     *
     * @yields {Scores} Snapshot at a point in time
     */
    * ScoresHistory() {
        let scores = new Scores()
        for (let award of this.PointsLog) {
            scores.Add(award)
            yield scores
        }
    }

    /**
     * Calculate the current scores.
     * 
     * @returns {Scores}
     */
    CurrentScores() {
        let scores
        for (scores of this.ScoreHistory());
        return scores
    }
}

/**
 * A MOTH Server interface.
 * 
 * This uses localStorage to remember Team ID,
 * and will send a Team ID with every request, if it can find one.
 */
class Server {
    /**
     * @param {string | URL} baseUrl Base URL to server, for constructing API URLs
     */
    constructor(baseUrl) {
        if (!baseUrl) {
            throw("Must provide baseURL")
        }
        this.baseUrl = new URL(baseUrl, location)
        this.teamIDKey = this.baseUrl.toString() + " teamID"
        this.TeamID = localStorage[this.teamIDKey]
    }

    /**
     * Fetch a MOTH resource.
     * 
     * If anything other than a 2xx code is returned,
     * this function throws an error.
     * 
     * This always sends teamID.
     * If args is set, POST will be used instead of GET
     * 
     * @param {string} path Path to API endpoint
     * @param {Object.<string,string>} args Key/Values to send in POST data
     * @returns {Promise.<Response>} Response
     */
    fetch(path, args={}) {
        let body = new URLSearchParams(args)
        if (this.TeamID && !body.has("id")) {
            body.set("id", this.TeamID)
        }

        let url = new URL(path, this.baseUrl)
        return fetch(url, {
            method: "POST",
            body,
            cache: "no-cache",
        })
    }

    /**
     * Send a request to a JSend API endpoint.
     * 
     * @param {string} path Path to API endpoint
     * @param {Object.<string,string>} args Key/Values to send in POST
     * @returns {Promise.<Object>} JSend Data
     */
    async call(path, args={}) {
        let resp = await this.fetch(path, args)
        let obj = await resp.json()
        switch (obj.status) {
            case "success":
                return obj.data
            case "fail":
                throw new Error(obj.data.description || obj.data.short || obj.data)
            case "error":
                throw new Error(obj.message)
            default:
                throw new Error(`Unknown JSend status: ${obj.status}`)
        }
    }

    /**
     * Make a new URL for the given resource.
     *
     * The returned URL instance will be absolute, and immune to changes to the
     * page that would affect relative URLs.
     *
     * @returns {URL}
     */
    URL(url) {
        return new URL(url, this.baseUrl)
    }

    /**
     * Are we logged in to the server?
     * 
     * @returns {boolean}
     */
    LoggedIn() {
        return this.TeamID ? true : false
    }

    /**
     * Forget about any previous Team ID.
     * 
     * This is equivalent to logging out.
     */
    Reset() {
        localStorage.removeItem(this.teamIDKey)
        this.TeamID = null
    }

    /**
     * Fetch current contest state.
     * 
     * @returns {Promise.<State>}
     */
    async GetState() {
        let resp = await this.fetch("/state")
        let obj = await resp.json()
        return new State(this, obj)
    }

    /**
     * Log in to a team.
     *
     * This calls the server's registration endpoint; if the call succeds, or
     * fails with "team already exists", the login is returned as successful. 
     *
     * @param {string} teamID
     * @param {string} teamName 
     * @returns {Promise.<string>} Success message from server
     */
    async Login(teamID, teamName) {
        let data = await this.call("/register", {id: teamID, name: teamName})
        this.TeamID = teamID
        this.TeamName = teamName
        localStorage[this.teamIDKey] = teamID
        return data.description || data.short
    }

    /**
     * Submit a proposed answer for points.
     *
     * The returned promise will fail if anything goes wrong, including the
     * proposed answer being rejected.
     *
     * @param {string} category Category of puzzle
     * @param {number} points Point value of puzzle
     * @param {string} proposed Answer to submit
     * @returns {Promise.<string>} Success message
     */
    async SubmitAnswer(category, points, proposed) {
        let data = await this.call("/answer", {
            cat: category, 
            points, 
            answer: proposed,
        })
        return data.description || data.short
    }

    /**
     * Fetch a file associated with a puzzle.
     * 
     * @param {string} category Category of puzzle
     * @param {number} points Point value of puzzle
     * @param {string} filename
     * @returns {Promise.<Response>}
     */
    GetContent(category, points, filename) {
        return this.fetch(`/content/${category}/${points}/${filename}`)
    }

    /**
     * Return a Puzzle object.
     * 
     * New Puzzle objects only know their category and point value.
     * See docstrings on the Puzzle object for more information.
     * 
     * @param {string} category 
     * @param {number} points 
     * @returns {Puzzle}
     */
    GetPuzzle(category, points) {
        return new Puzzle(this, category, points)
    }
}

export {
    Hash,
    Server,
}