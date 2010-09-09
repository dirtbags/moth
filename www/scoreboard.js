function dbg(o) {
    e = document.getElementById("debug");
    e.innerHTML = o;
}

function torgba(color, alpha) {
    if (color.substring(0, 1) == "#") {
        var r = parseInt(color.substring(1,3), 16);
        var g = parseInt(color.substring(3,5), 16);
        var b = parseInt(color.substring(5,7), 16);

        return "rgba(" + r + "," + g + "," + b + "," + alpha + ")";
    } else {
        return color;
    }
}

function Chart(id, width, height, lines) {
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

    function draw(color, values) {
        var lasty = 0;

        ctx.strokeStyle = color;
        ctx.lineWidth = 4;
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

    this.highlight = function(id, color) {
        var line = lines[id];
        if (! color) color = line[0];

        draw(color, line[1]);
    }

    for (id in lines) {
        var line = lines[id];

        draw(line[0], line[1]);
    }

}

var thechart;

function plot(id, width, height, lines) {
    thechart = new Chart(id, width, height, lines);
}

function getElementsByClass( searchClass, domNode, tagName) {
    if (domNode == null) domNode = document;
    if (tagName == null) tagName = '*';
    var el = new Array();
    var tags = domNode.getElementsByTagName(tagName);
    var tcl = " "+searchClass+" ";
    for(i=0,j=0; i<tags.length; i++) {
        var test = " " + tags[i].className + " ";
        if (test.indexOf(tcl) != -1)
            el[j++] = tags[i];
    }
    return el;
}

function highlight(cls, color) {
    if (! color) color = "#ffffff";
    elements = getElementsByClass("t" + cls);
    for (i in elements) {
        e = elements[i];
        e.style.borderColor = e.style.backgroundColor;
        e.style.backgroundColor = color;
    }
    thechart.highlight(cls, color);
}

function restore(cls) {
    elements = getElementsByClass("t" + cls);
    for (i in elements) {
        e = elements[i];
        e.style.backgroundColor = e.style.borderColor;
    }
    thechart.highlight(cls);
}


