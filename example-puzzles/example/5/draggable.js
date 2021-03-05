// jshint asi:true

var dragSrcEl_

function draggableHandleDragStart(e) {
  e.target.dataset.moveId = e.timeStamp.toString()
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', e.target.dataset.moveId)

  // this/e.target is the source node.
  e.target.classList.add('moving')
}

function draggableHandleDragOver(e) {
  if (e.target.attributes.draggable) {
    e.preventDefault() // Allows us to drop.
  }

  e.dataTransfer.dropEffect = 'move'

  return false
}

function draggableHandleDragEnter(e) {
  let element = e.target
  if (!element.classList) {
    element = element.parentElement
  }
  element.classList.add('over')
}

function draggableHandleDragLeave(e) {
  // this/e.target is previous target element.
  let element = e.target
  if (!element.classList) {
    element = element.parentElement
  }
  element.classList.remove('over')
}

function draggableHandleDrop(e) {
  // this/e.target is current target element.
  e.preventDefault()
  let tgt = e.target
  let src = document.querySelector("[data-move-id=\"" + e.dataTransfer.getData("text/plain") + "\"]")

  // Don't do anything if we're dropping on the same column we're dragging.
  if (src == tgt) {
    return false
  }
  
  let tgtPrev = tgt.previousSibling
  src.replaceWith(tgt)
  tgtPrev.after(src)
  
  tgt.dispatchEvent(new InputEvent("input", {bubbles: true}))
}

function draggableHandleDragEnd(e) {
  // this/e.target is the source node.
  for (e of document.querySelectorAll("[draggable].over")) {
    e.classList.remove("over")
  }
  for (e of document.querySelectorAll("[draggable].moving")) {
    e.classList.remove("moving")
  }
}

function sortableInit(event) {
  for (let e of document.querySelectorAll("[draggable]")) {
    e.addEventListener('dragstart', draggableHandleDragStart)
    e.addEventListener('dragenter', draggableHandleDragEnter)
    e.addEventListener('dragover', draggableHandleDragOver)
    e.addEventListener('dragleave', draggableHandleDragLeave)
    e.addEventListener('drop', draggableHandleDrop)
    e.addEventListener('dragend', draggableHandleDragEnd)
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", sortableInit)
} else {
  sortableInit()
}
