function Plot(id, width, height) {
    var canvas = document.getElementById(id);
    var ctx = canvas.getContext('2d');

    canvas.width = 800;
    canvas.height = 200;

    // We'll let the canvas do all the tricksy math
    xscale = canvas.width/width;
    yscale = canvas.height/height;
    ctx.lineWidth = 2;

    function moveTo(x, y) {
        ctx.moveTo(Math.round(x * xscale), Math.round(y * yscale));
    }
    function lineTo(x, y) {
        ctx.lineTo(Math.round(x * xscale), Math.round(y * yscale));
    }

    function draw(values) {
        ctx.beginPath();
        moveTo(values[0][0], height);
        var lasty = 0;
        for (i in values) {
            var x = values[i][0];
            var y = values[i][1];
            lineTo(x, height - lasty);
            lineTo(x, height - y);
            lasty = y;
        }
        lineTo(width, height - lasty);
    }


    this.line = function(color, values) {
        ctx.fillStyle = color;
        ctx.strokeStyle = color;

        draw(values);
        ctx.stroke();
    }
}
