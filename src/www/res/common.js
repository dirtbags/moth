function loadJSON(url, callback) {
    function loaded(e) {
	callback(e.target.response);
    }
    var xhr = new XMLHttpRequest()
    xhr.onload = loaded;
    xhr.open("GET", url, true);
    xhr.responseType = "json";
    xhr.send();
}

function createElement(tagName) {
    return document.createElement(tagName);
}

function djb2hash(str) {
    var hash = 5381;

    for (var i = 0; i < str.length; i += 1) {
	var c = str.charCodeAt(i);
	hash = ((hash * 33) + c) & 0xffffffff;
    }
    return hash;
}

// Make code readable by providing Function.prototype.bind in older JS environments
if (!Function.prototype.bind) {
  Function.prototype.bind = function(oThis) {
    if (typeof this !== 'function') {
      // closest thing possible to the ECMAScript 5
      // internal IsCallable function
      throw new TypeError('Function.prototype.bind - what is trying to be bound is not callable');
    }

    var aArgs   = Array.prototype.slice.call(arguments, 1),
        fToBind = this,
        fNOP    = function() {},
        fBound  = function() {
          return fToBind.apply(this instanceof fNOP && oThis
                 ? this
                 : oThis,
                 aArgs.concat(Array.prototype.slice.call(arguments)));
        };

    fNOP.prototype = this.prototype;
    fBound.prototype = new fNOP();

    return fBound;
  };
}
