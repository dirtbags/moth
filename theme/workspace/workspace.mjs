import {Toast} from "../common.mjs"
import "https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"
import * as CodeJar from "https://cdn.jsdelivr.net/npm/codejar@4.2.0"

var workers = {}

// loadWorker returns an existing worker if one exists, otherwise, it starts a new worker
function loadWorker(language) {
    let worker = workers[language]
    if (!worker) {
        let url = new URL(language + ".mjs", import.meta.url)
        worker = new Worker(url, {
            type: "module",
        })
        workers[language] = worker
    }
    return worker
}

export class Workspace {
    /**
     * 
     * @param codeBlock {HTMLElement} The element containing the source code
     * @param id {string} A unique identifier of this workspace
     * @param attachmentUrls {URL[]} List of attachment URLs
     */
    constructor(codeBlock, id, attachmentUrls) {
        // Show a progress bar
        let loadingElement = document.createElement("progress")
        codeBlock.insertAdjacentElement("afterend", loadingElement)

        this.language = "unknown"
        for (let c of codeBlock.classList) {
            let parts = c.split("-")
            if ((parts.length == 2) && parts[0].startsWith("lang")) {
                this.language = parts[1]
            }
        }
    
        this.element = document.createElement("div")
        this.element.classList.add("workspace")
        let template = document.querySelector("template#workspace")
        this.element.appendChild(template.content.cloneNode(true))
    
        this.originalCode = codeBlock.textContent
        this.attachmentUrls = attachmentUrls
        this.storageKey = "code:" + id

        // Get our document and window
        this.document = this.element.ownerDocument
        this.window = this.document.defaultView
    
        // Load user modifications, if there are any
        this.code = localStorage[this.storageKey] || this.originalCode
    
        this.status = this.element.querySelector(".status")
        this.linenos = this.element.querySelector(".editor .linenos")
        this.editor = this.element.querySelector(".editor .text")
        this.stdout = this.element.querySelector(".stdout")
        this.stderr = this.element.querySelector(".stderr")
        this.traceback = this.element.querySelector(".traceback")
        this.stdinfo = this.element.querySelector(".stdinfo")
        this.runButton = this.element.querySelector("button.run")
        this.revertButton = this.element.querySelector("button.revert")
        this.fontButton = this.element.querySelector("button.font")
        this.element.querySelector(".language").textContent = this.language

        this.runButton.disabled = true
    
        // Load in the editor
        this.editor.classList.add("language-" + this.language)
        this.jar = CodeJar.CodeJar(this.editor, (editor) => this.highlight(editor), {window: this.window})
        this.jar.updateCode(this.code)
        switch (this.language) {
            case "python":
                this.jar.updateOptions({
                    tab: "    ",
                    indentOn: /:$/,
                })
                break
        }

        // Load the interpreter
        this.initLanguage(this.language)
        .then(() => {
            codeBlock.parentElement.replaceWith(this.element)
        })
        .catch(err => console.warn(`Unable to load ${this.language} interpreter`))
        .finally(() => {
            loadingElement.remove()
        })
        this.runButton.addEventListener("click", () => this.run())
        this.revertButton.addEventListener("click", () => this.revert())
        this.fontButton.addEventListener("click", () => this.font())

    }

    initLanguage(language) {
        let start = performance.now()
        this.status.textContent = "Initializing..."
        this.status.appendChild(document.createElement("progress"))

        let workerUrl = new URL(language + ".mjs", import.meta.url)
        this.worker = new Worker(workerUrl, {type: "module"})

        // XXX: There has got to be a cleaner way to do this
        return new Promise((resolve, reject) => {
            this.worker.addEventListener("error", err => reject(err))
            this.workerMessage({type: "nop"})
            .then(() => {
                let runtime = performance.now() - start
                let duration = new Date(runtime).toISOString().slice(11, -1)        
                this.status.textContent = "Loaded in " + duration
                this.runButton.disabled = false
        
                for (let a of this.attachmentUrls) {
                    let filename = a.pathname.split("/").pop()
                    this.workerMessage({type: "wget", url: a.href || a})
                    .then(ret => {
                        this.stdinfo.appendChild(this.document.createElement("div")).textContent = "Downloaded " + filename
                    })
                }
                resolve()
            })
        })
    }

    workerMessage(message) {
        let chan = new MessageChannel()
        message.channel = chan.port2
        this.worker.postMessage(message, [chan.port2])
        let p = new Promise(
            (resolve, reject) => {
                chan.port1.addEventListener("message", e => resolve(e.data), {once: true})
            }
        )
        chan.port1.start()
        return p
    }

    workerReady() {
        return this.workerMessage({type: "nop"})
    }

    workerWget(url) {
        return this.workerMessage({
            type: "wget",
            url: url.href || url,
        })
    }

    /**
     * highlight provides a code highlighter for CodeJar
     * 
     * It calls Prism.highlightElement, then updates line numbers
     */
    highlight(editor) {
        if (Prism) {
            // Sometimes it loads slowly
            Prism.highlightElement(editor)
        } else {
            console.warn("No highlighter!", Prism, this.window.document.scripts)
        }

        // Create a line numbers column
        if (true) {
            const code = editor.textContent || ""
            const lines = code.split("\n")
            let linesCount = lines.length
            if (lines[linesCount-1]) {
                linesCount += 1
            }
    
            let ltxt = ""
            for (let i = 1; i < linesCount; i++) {
                ltxt += i + "\n"
            }
            this.linenos.textContent = ltxt
        }
    }

    setAnswer(answer) {
        let evt = new CustomEvent("setAnswer", {detail: {value: answer}, bubbles: true, cancelable: true})
        this.element.dispatchEvent(evt)

        this.stdinfo.appendChild(this.document.createTextNode("Set answer to "))
        this.stdinfo.appendChild(this.document.createElement("code")).textContent = answer
    }

    async run() {
        let start = performance.now()
        this.runButton.disabled = true
        this.status.textContent = "Running..."
        
        // Save first. Always save first.
        let program = this.jar.toString()
        if (program != this.originalCode) {
            localStorage[this.storageKey] = program
        }
        
        let result = await this.workerMessage({
            type: "run",
            code: program,
        })
        
        this.stdout.textContent = result.stdout
        this.stderr.textContent = result.stderr
        this.traceback.textContent = result.traceback
        while (this.stdinfo.firstChild) this.stdinfo.firstChild.remove()
        if (result.answer) {
            this.setAnswer(result.answer)
        }

        let runtime = performance.now() - start
        let duration = new Date(runtime).toISOString().slice(11, -1)
        this.status.textContent = "Ran in " + duration
        this.runButton.disabled = false
    }
    
    revert() {
        let currentCode = this.jar.toString()
        let savedCode = localStorage[this.storageKey]
        if ((currentCode == this.originalCode) && savedCode) {
            this.jar.updateCode(savedCode)
            Toast("Re-loaded saved code")
        } else {
            this.jar.updateCode(this.originalCode)
            Toast("Reverted to original code")
        }
    }

    font(force) {
        this.element.classList.toggle("fixed", force)
    }
}
