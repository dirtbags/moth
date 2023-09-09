function randint(max) {
    return Math.floor(Math.random() * max)
}

const MILLISECOND = 1
const SECOND = MILLISECOND * 1000

class Line {
    /**
     * @param {CanvasRenderingContext2D} ctx canvas context
     * @param {Number} hue Hue, in % of one circle [0,tau)
     * @param {Number} a First point of line
     * @param {Number} b Second point of line
     */
    constructor(ctx, hue, a, b) {
        this.ctx = ctx
        this.hue = hue
        this.a = a
        this.b = b
    }

    bounce(point, v) {
        let ret = [
            point[0] + v[0],
            point[1] + v[1],
        ]
        if ((ret[0] > this.ctx.canvas.width) || (ret[0] < 0)) {
            v[0] *= -1
            ret[0] += v[0] * 2
        }
        if ((ret[1] > this.ctx.canvas.height) || (ret[1] < 0)) {
            v[1] *= -1
            ret[1] += v[1] * 2
        }
        return ret
    }

    Add(hue, a, b) {
        return new Line(
            this.ctx,
            (this.hue + hue) % 1.0,
            this.bounce(this.a, a),
            this.bounce(this.b, b),
        )
    }

    Draw() {
        this.ctx.save()
        this.ctx.strokeStyle = `hwb(${this.hue}turn 0% 50%)`
        this.ctx.beginPath()
        this.ctx.moveTo(this.a[0], this.a[1])
        this.ctx.lineTo(this.b[0], this.b[1])
        this.ctx.stroke()
        this.ctx.restore()
    }
}

class LengoBackground {
    constructor() {
        this.canvas = document.createElement("canvas")
        document.body.insertBefore(this.canvas, document.body.firstChild)
        this.canvas.style.position = "fixed"
        this.canvas.style.zIndex = -1000
        this.canvas.style.opacity = 0.3
        this.canvas.style.top = 0
        this.canvas.style.left = 0
        this.canvas.style.width = "99vw"
        this.canvas.style.height = "99vh"
        this.canvas.width = 2000
        this.canvas.height = 2000
        this.ctx = this.canvas.getContext("2d")
        this.ctx.lineWidth = 1

        this.lines = []
        for (let i = 0; i < 18; i++) {
            this.lines.push(
                new Line(this.ctx, 0, [0, 0], [0, 0])
            )
        }
        this.velocities = {
            hue: 0.001,
            a: [20 + randint(10), 20 + randint(10)],
            b: [5 + randint(10), 5 + randint(10)],
        }
        this.nextFrame = performance.now()-1
        this.frameInterval = 100 * MILLISECOND

        //addEventListener("resize", e => this.resizeEvent())
        //this.resizeEvent()
        //this.animate(this.nextFrame)
        setInterval(() => this.animate(this.nextFrame+1), SECOND/6)
    }


    /**
     * Animate one frame
     * 
     * @param {DOMHighResTimeStamp} timestamp 
     */
    animate(timestamp) {
        if (timestamp >= this.nextFrame) {
            this.lines.shift()
            let lastLine = this.lines.pop()
            let nextLine = lastLine.Add(this.velocities.hue, this.velocities.a, this.velocities.b)
            this.lines.push(lastLine)
            this.lines.push(nextLine)

            this.ctx.clearRect(0, 0, this.ctx.canvas.width, this.ctx.canvas.height)
            for (let line of this.lines) {
                line.Draw()
            }
            this.nextFrame += this.frameInterval
        }
        //requestAnimationFrame((ts) => this.animate(ts))
    }
}

function init() {
    new LengoBackground()
}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
