function torgba(color, alpha) {
    var r = parseInt(color.substring(1,3), 16);
    var g = parseInt(color.substring(3,5), 16);
    var b = parseInt(color.substring(5,7), 16);

    return "rgba(" + r + "," + g + "," + b + "," + alpha + ")";
}

function start(turns) {
    var canvas = document.getElementById('battlefield');
    var ctx = canvas.getContext('2d');
    var loop_id;


    function crater(color, x, y, rotation) {
        var points = 7;
        var angle = Math.PI / points;

        ctx.save();
        ctx.translate(x, y);
        ctx.rotate(rotation);
        ctx.lineWidth = 2;
        ctx.strokeStyle = torgba(color, 0.5);
        ctx.fillStyle = torgba(color, 0.2);
        ctx.beginPath();
        ctx.moveTo(12, 0);
        for (i = 0; i < points; i += 1) {
            ctx.rotate(angle);
            ctx.lineTo(6, 0);
            ctx.rotate(angle);
            ctx.lineTo(12, 0);
        }
        ctx.closePath()
        ctx.stroke();
        ctx.fill();
        ctx.restore();
    }

    function sensors(color, x, y, rotation, sensors) {
        var sensor_color = torgba(color, 0.4);
        ctx.save();
        ctx.translate(x, y);
        ctx.rotate(rotation);
        ctx.lineWidth = 1;
        for (i in sensors) {
            s = sensors[i];
            if (s[3]) {
                ctx.strokeStyle = "#000";
            } else {
                ctx.strokeStyle = sensor_color;
            }
            ctx.beginPath();
            ctx.moveTo(0, 0);
            ctx.arc(0, 0, s[0], s[1], s[2], false);
            ctx.closePath();
            ctx.stroke();
        }
        ctx.restore();
    }

    function tank(color, x, y, rotation, turret, led, fire) {
        ctx.save();
        ctx.fillStyle = color;
        ctx.translate(x, y);
        ctx.rotate(rotation);
        ctx.fillRect(-5, -4, 10, 8);
        ctx.fillStyle = "#777777";
        ctx.fillRect(-7, -9, 15, 5);
        ctx.fillRect(-7,  4, 15, 5);
        ctx.rotate(turret);
        if (fire) {
            ctx.fillStyle = color;
            ctx.fillRect(0, -1, fire, 2);
        } else {
            if (led) {
                ctx.fillStyle = "#ff0000";
            } else {
                ctx.fillStyle = "#000000";
            }
            ctx.fillRect(0, -1, 10, 2);
        }
        ctx.restore();
    }

    var frame = 0;
    var lastframe = 0;
    var fps = document.getElementById('fps');
    function update_fps() {
        fps.innerHTML = (frame - lastframe);
        lastframe = frame;
    }
    function update() {
        var idx = frame % (turns.length + 20);

        frame += 1;
        if (idx >= turns.length) {
            return;
        }

        canvas.width = canvas.width;
        turn = turns[idx];

        // Draw craters first
        for (i in turn) {
            t = turn[i];
            if (t[0]) {
                crater(t[1], t[2], t[3], t[4]);
            }
        }
        // Then sensors
        for (i in turn) {
            t = turn[i];
            if (! t[0]) {
                sensors(t[1], t[2], t[3], t[4], t[8]);
            }
        }
        // Then tanks
        for (i in turn) {
            t = turn[i];
            if (! t[0]) {
                // Surely there's a better way.  CBA right now.
                tank(t[1], t[2], t[3], t[4], t[5], t[6], t[7]);
            }
        }
    }

    loop_id = setInterval(update, 66);
    setInterval(update_fps, 1000);
}

