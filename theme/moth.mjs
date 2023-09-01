class Server {
    constructor(baseUrl) {
        this.baseUrl = new URL(baseUrl)
        this.teamId = null
    }

    /**
     * Fetch a MOTH resource.
     * 
     * This is just a convenience wrapper to always send teamId.
     * If body is set, POST will be used instead of GET
     * 
     * @param {String} path Path to API endpoint
     * @param {Object} body Key/Values to send in POST data
     * @returns {Promise} Response
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
     * @param {Object} args Key/Values to send in POST
     * @returns JSend Data
     */
    async postJSend(path, args) {
        let resp = await this.fetch(path, args)
        if (!resp.ok) {
            throw new Error(resp.statusText)
        }
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
     * Register a team name with a team ID.
     * 
     * This is similar to, but not exactly the same as, logging in.
     * See MOTH documentation for details.
     * 
     * @param {String} teamId 
     * @param {String} teamName 
     * @returns {String} Success message from server
     */
    async Register(teamId, teamName) {
        let data = await postJSend("/login", {id: teamId, name: teamName})
        this.teamId = teamId
        this.teamName = teamName
        return data.description || data.short
    }

    /**
     * Fetch current contest status.
     * 
     * @returns {Object} Contest status
     */
    async Status() {
        let data = await this.postJSend("/status")
        return data
    }
}

export {
    Server
}