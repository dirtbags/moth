// jshint asi:true

async function helperUpdateAnswer(event) {
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
    if (values.length == 0) {
      value = "None"
    } else {
      value = values.join(join)
    }
  }

  // First make any adjustments to the value
  if (e.classList.contains("lower")) {
    value = value.toLowerCase()
  }
  if (e.classList.contains("upper")) {
    value = value.toUpperCase()
  }

  // "substrings" answers try all substrings. If any are the answer, they're filled in.
  if (e.classList.contains("substring")) {
    let validated = null
    let anchorEnd = e.classList.contains("anchor-end")
    let anchorBeg = e.classList.contains("anchor-beg")

    for (let end = 0; end <= value.length; end += 1) {
      for (let beg = 0; beg < value.length; beg += 1) {
        if (anchorEnd && (end != value.length)) {
          continue
        }
        if (anchorBeg && (beg != 0)) {
          continue
        }
        let sub = value.substring(beg, end)
        if (await checkAnswer(sub)) {
          validated = sub
        }
      }
    }

    value = validated
  }

  // If anything zeroed out value, don't update the answer field
  if (!value) {
    return
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

{
  let init = function(event) {
    for (let e of document.querySelectorAll(".answer")) {
      helperActivate(e)
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init)
  } else {
    init()
  }
}
