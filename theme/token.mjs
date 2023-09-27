/**
 * Functionality for token.html
 */
import * as moth from "./moth.mjs"
import * as common from "./common.mjs"

const server = new moth.Server(".")

/**
 * Handle a submit event on a form.
 * 
 * @param {SubmitEvent} event 
 */
async function formSubmitHandler(event) {
    event.preventDefault()

    let formData = new FormData(event.target)
    let token = formData.get("token")
    let vals = token.split(":")
    let category = vals[0]
    let points = Number(vals[1])
    let proposed = vals[2]
    if (!category || !points || !proposed) {
        console.info("Not a token:", vals)
        common.Toast("This is not a properly-formed token")
        return
    }
    try {
        let message = await server.SubmitAnswer(category, points, proposed)
        common.Toast(message)
    }
    catch (error) {
        if (error.message == "incorrect answer") {
            common.Toast("Unknown token")
        } else {
            console.error(error)
            common.Toast(error)
        }
    }
}

function init() {
    for (let form of document.querySelectorAll("form.token")) {
        form.addEventListener("submit", formSubmitHandler)
    }
}

common.WhenDOMLoaded(init)