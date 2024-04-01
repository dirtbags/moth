import * as pyodide from "https://cdn.jsdelivr.net/npm/pyodide@0.25.1/pyodide.mjs" // v0.16.1 known good

const HOME = "/home/web_user"

async function createInstance() {
    let instance = await pyodide.loadPyodide()
    instance.runPython("import sys")
    self.postMessage({type: "loaded"})
    return instance
}
const initialized = createInstance()

class Buffer {
    constructor() {
        this.buf = []
    }

    write(s) {
        this.buf.push(s)
    }

    value() {
        return this.buf.join("")
    }
}

async function handleMessage(event) {
    let data = event.data
    
    let instance = await initialized
    let fs = instance._module.FS

    let ret = {
        result: null,
        answer: null,
        stdout: null,
        stderr: null,
        traceback: null,
    }

    switch (data.type) {
        case "nop":
            // You might want to do nothing in order to display to the user that a run can now be handled
            break
        case "run":
            let sys = instance.globals.get("sys")
            sys.stdout = new Buffer()
            sys.stderr = new Buffer()
            instance.globals.set("setanswer", (s) => {ret.answer = s})

            try {
                ret.result = await instance.runPythonAsync(data.code)
            } catch (err) {
                ret.traceback = err
            }
            ret.stdout = sys.stdout.value()
            ret.stderr = sys.stderr.value()
            break
        case "wget":
            let url = data.url
            let dir = data.directory || fs.cwd()
            let filename = url.split("/").pop()
            let path = dir + "/" + filename

            if (fs.analyzePath(path).exists) {
                fs.unlink(path)
            }
            fs.createLazyFile(dir, filename, url, true, false)
            break
        default:
            ret.result = "Unknown message type: " + data.type
            break
    }
    if (data.channel) {
        data.channel.postMessage(ret)
    }
}
self.addEventListener("message", e => handleMessage(e))
