// jshint asi:true

function helperUpdateAnswer(event) {
  let e = event.currentTarget
  let value = e.value
  let inputs = e.querySelectorAll("input")
  
  if (inputs.length > 0) {
    // If there are child input nodes, join their values with commas
    let values = []
    for (let c of inputs) {
      if (c.type == "checkbox") {
        if (c.checked) {
          values.push(c.value)
        }
      } else {
        values.push(c.value)
      }
    }
    if (e.classList.contains("sort")) {
      values.sort()
    }
    let join = e.dataset.join
    if (join === undefined) {
      join = ","
    }
    value = values.join(join)
  }

  // First make any adjustments to the value
  if (e.classList.contains("lower")) {
    value = value.toLowerCase()
  }
  if (e.classList.contains("upper")) {
    value = value.toUpperCase()
  }

  let answer = document.querySelector("#answer")
  answer.value = value
  answer.dispatchEvent(new InputEvent("input"))
}

function helperRemoveInput(e) {
  let item = e.target.parentElement
  let container = item.parentElement
  item.remove()
  
  var event = new Event("input")
  container.dispatchEvent(event)
}

function helperExpandInputs(e) {
  let item = e.target.parentElement
  let container = item.parentElement
  let template = container.firstElementChild
  let newElement = template.cloneNode(true)

  // Add remove button
  let remove = document.createElement("button")
  remove.innerText = "➖"
  remove.title = "Remove this input"
  remove.addEventListener("click", helperRemoveInput)
  newElement.appendChild(remove)

  // Zero it out, otherwise whatever's in first element is copied too
  newElement.querySelector("input").value = ""

  container.insertBefore(newElement, item)
  
  var event = new Event("input")
  container.dispatchEvent(event)
}

function helperActivate(e) {
  e.addEventListener("input", helperUpdateAnswer)
  for (let exp of e.querySelectorAll(".expand")) {
    exp.addEventListener("click", helperExpandInputs)
  }
}

function helperInit(event) {
  for (let e of document.querySelectorAll(".answer")) {
    helperActivate(e)
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", helperInit);
} else {
  helperInit();
}
