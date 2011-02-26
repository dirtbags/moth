sessid = "";

function go() {
    input = document.getElementById("input");
    output = document.getElementById("output");
    prompt = document.getElementById("prompt");
    val = input.value;

    r = new XMLHttpRequest();

    function statechange() {
        if (r.readyState == 4) {
            doc = r.responseXML;
            resp_txt = doc.getElementsByTagName("response")[0].textContent;
            prompt_txt = doc.getElementsByTagName("prompt")[0].textContent;
            error = doc.getElementsByTagName("error")[0];

            if (! sessid) {
                sessid = doc.getElementsByTagName("sessionid")[0].textContent;
                output.textContent += "Connected with session ID " + sessid + ".\n";
            }

            if (resp_txt) {
                if (resp_txt.charAt(resp_txt.length-1) != "\n") {
                    resp_txt += "\n";
                }
                output.textContent += resp_txt;
            }
            if (error) {
                e = document.createElement('div');
                e.className = 'error';
                e.textContent = error.textContent + "\n";
                output.appendChild(e);
            }
            if (prompt_txt) {
                prompt.textContent = prompt_txt;
            }

            prompt.style.display = "inline";
        }
        document.body.scrollTop = document.body.scrollHeight;
        input.focus();
    }

    // Calculate this before screwing with stuff
    data = ("s=" + sessid + '&v=' + escape(val));

    // Add prompt and input text to output.  This instantly displays the
    // text so you know you hit enter, while providing a slight delay in
    // results, like the server is "working on" the request.
    output.textContent += prompt.textContent + " " + val + "\n";
    input.value = "";
    input.focus();              // Maybe prevent color flashes

    setTimeout(statechange, 1);
    if (val == "@sessid") {
        output.textContent += sessid + "\n";
    } else if (val[0] == ":") {
        code = val.substr(1, val.length - 1);
        output.textContent += "==> " + eval(code) + "\n";
    } else {
        prompt.style.display = "none";
        r.onreadystatechange = statechange;
        r.open("POST", "wopr.cgi");

        r.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        r.setRequestHeader("Content-length", data.length);
        r.send(data);
    }

    return false;
}

window.onload = go;
