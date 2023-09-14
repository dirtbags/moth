/**
 * Common functionality
 */
const Millisecond = 1
const Second = Millisecond * 1000
const Minute = Second * 60

/**
 * Display a transient message to the user.
 * 
 * @param {String} message Message to display
 * @param {Number} timeout How long before removing this message
 */
function Toast(message, timeout=5*Second) {
    console.info(message)
    for (let toasts of document.querySelectorAll(".toasts")) {
        let p = toasts.appendChild(document.createElement("p"))
        p.classList.add("toast")
        p.textContent = message
        setTimeout(() => p.remove(), timeout)
    }
}

/**
 * Run a function when the DOM has been loaded.
 * 
 * @param {function():void} cb Callback function
 */
function WhenDOMLoaded(cb) {
    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", cb)
    } else {
        cb()
    }    
}

export {
    Millisecond,
    Second,
    Minute,
    Toast,
    WhenDOMLoaded,
}
