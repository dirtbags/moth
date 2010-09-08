function dbg(o) {
    e = document.getElementById("debug");
    e.innerHTML = o;
}

function torgba(color, alpha) {
    var r = parseInt(color.substring(1,3), 16);
    var g = parseInt(color.substring(3,5), 16);
    var b = parseInt(color.substring(5,7), 16);

    return "rgba(" + r + "," + g + "," + b + "," + alpha + ")";
}

function plot(id, width, height, lines) {
    var canvas = document.getElementById(id);
    var ctx = canvas.getContext('2d');

    // We'll let the canvas do all the tricksy math
    var xscale = canvas.width/width;
    var yscale = canvas.height/height;
    var nlines = lines.length;

    function moveTo(x, y) {
        ctx.moveTo(Math.round(x * xscale), Math.round((height - y) * yscale));
    }
    function lineTo(x, y) {
        ctx.lineTo(Math.round(x * xscale), Math.round((height - y) * yscale));
    }

    function draw(line) {
        var color = line[0];
        var values = line[1];
        var lasty = 0;

        ctx.strokeStyle = torgba(color, 0.99);
        ctx.lineWidth = 2;
        ctx.beginPath();
        moveTo(values[0][0], 0);
        for (i in values) {
            var x = values[i][0];
            var y = values[i][1];
            lineTo(x, lasty);
            lineTo(x, y);
            lasty = y;
        }
        lineTo(width, lasty);
        ctx.stroke();
    }


    var cur = 0;

    function update() {
        var line = lines[cur];

        draw(line);
        cur = (cur + 1) % nlines;

        if (cur > 0) {
            setTimeout(update, 66);
        }
    }


    update()
}
