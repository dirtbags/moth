// Devel server addons

// devel_addin drops a bunch of development extensions into element e.
// It will only modify stuff inside e.
function devel_addin(obj, e) {
  let h = document.createElement("h2");
  e.appendChild(h);
  h.textContent = "Development Options";

  let g = document.createElement("p");
  e.appendChild(g);
  g.innerText = "This section will not appear for participants."
  
  let keys = Object.keys(obj);
  keys.sort();
  for (let key of keys) {
    switch (key) {
      case "body":
        continue;
    }

    let d = document.createElement("div");
    e.appendChild(d);
    d.classList.add("kvpair");
    
    let ktxt = document.createElement("span");
    d.appendChild(ktxt);
    ktxt.textContent = key;
    
    let val = obj[key];
    if (Array.isArray(val)) {
      let vi = document.createElement("select");
      d.appendChild(vi);
      vi.multiple = true;
      for (let a of val) {
        let opt = document.createElement("option");
        vi.appendChild(opt);
        opt.innerText = a;
      }
    } else {
      let vi = document.createElement("input");
      d.appendChild(vi);
      vi.value = val;
      vi.disabled = true;
    }
  }
}