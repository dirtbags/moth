function randint(max) {
    return Math.floor(Math.random() * max)
}

const MILLISECOND = 1
const SECOND = MILLISECOND * 1000

class Point {
    constructor(x, y) {
        this.x = x
        this.y = y
    }

    /**
     * Add n to this.
     * 
     * @param {Point} n What to add to this
     * @returns {Point}
     */
    Add(n) {
        return new Point(this.x + n.x, this.y + n.y)
    }

    /**
     * Subtract n from this.
     * 
     * @param {Point} n 
     * @returns {Point}
     */
    Subtract(n) {
        return new Point(this.x - n.x, this.y - n.y)
    }

    /**
     * Add velocity, then bounce point off box defined by points at min and max
     * @param {Point} velocity
     * @param {Point} min 
     * @param {Point} max 
     * @returns {Point}
     */
    Bounce(velocity, min, max) {
        let p = this.Add(velocity)
        if (p.x < min.x) {
            p.x += (min.x - p.x) * 2
            velocity.x *= -1
        }
        if (p.x > max.x) {
            p.x += (max.x - p.x) * 2
            velocity.x *= -1
        }
        if (p.y < min.y) {
            p.y += (min.y - p.y) * 2
            velocity.y *= -1
        }
        if (p.y > max.y) {
            p.y += (max.y - p.y) * 2
            velocity.y *= -1
        }
        return p
    }

    /**
     * 
     * @param {Point} p 
     * @returns {Boolean}
     */
    Equal(p) {
        return (this.x == p.x) && (this.y == p.y)
    }
}

class QixLine {
    /**
     * @param {Number} hue 
     * @param {Point} a 
     * @param {Point} b 
     */
    constructor(hue, a, b) {
        this.hue = hue
        this.a = a
        this.b = b
    }
}

/**
 * Draw a line dancing around the screen,
 * like the video game "qix"
 */
class QixBackground {
    constructor(ctx, frameInterval = SECOND/6) {
        this.ctx = ctx
        this.min = new Point(0, 0)
        this.max = new Point(this.ctx.canvas.width, this.ctx.canvas.height)
        this.box = this.max.Subtract(this.min)

        this.lines = [
            new QixLine(
                0,
                new Point(randint(this.box.x), randint(this.box.y)),
                new Point(randint(this.box.x), randint(this.box.y)),
            )
        ]
        while (this.lines.length < 18) {
            this.lines.push(this.lines[0])
        }
        this.velocity = new QixLine(
            0.001,
            new Point(1 + randint(this.box.x / 100), 1 + randint(this.box.y / 100)),
            new Point(1 + randint(this.box.x / 100), 1 + randint(this.box.y / 100)),
        )

        this.frameInterval = frameInterval
        this.nextFrame = 0
    }

    /**
     * Maybe draw a frame
     */
    Animate() {
        let now = performance.now()
        if (now < this.nextFrame) {
            // Not today, satan
            return
        }
        this.nextFrame = now + this.frameInterval

        this.lines.shift()
        let lastLine = this.lines[this.lines.length - 1]
        let nextLine = new QixLine(
            (lastLine.hue + this.velocity.hue) % 1.0,
            lastLine.a.Bounce(this.velocity.a, this.min, this.max),
            lastLine.b.Bounce(this.velocity.b, this.min, this.max),
        )

        this.lines.push(nextLine)

        this.ctx.clearRect(0, 0, this.ctx.canvas.width, this.ctx.canvas.height)
        for (let line of this.lines) {
            this.ctx.save()
            this.ctx.strokeStyle = `hwb(${line.hue}turn 0% 0%)`
            this.ctx.beginPath()
            this.ctx.moveTo(line.a.x, line.a.y)
            this.ctx.lineTo(line.b.x, line.b.y)
            this.ctx.stroke()
            this.ctx.restore()
        }
    }
}

function init() {
    let canvas = document.createElement("canvas")
    canvas.width = 640
    canvas.height = 640
    canvas.classList.add("wallpaper")
    document.body.insertBefore(canvas, document.body.firstChild)

    let ctx = canvas.getContext("2d")

    let qix = new QixBackground(ctx)
    setInterval(() => qix.Animate(), SECOND/6)

}

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
} else {
    init()
}
