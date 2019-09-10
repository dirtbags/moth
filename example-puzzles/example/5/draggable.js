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
  e.target.classList.add('over')
}

function draggableHandleDragLeave(e) {
  // this/e.target is previous target element.
  e.target.classList.remove('over')
}

function draggableHandleDrop(e) {
  // this/e.target is current target element.
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
    e.addEventListener('dragstart', draggableHandleDragStart, false)
    e.addEventListener('dragenter', draggableHandleDragEnter, false)
    e.addEventListener('dragover', draggableHandleDragOver, false)
    e.addEventListener('dragleave', draggableHandleDragLeave, false)
    e.addEventListener('drop', draggableHandleDrop, false)
    e.addEventListener('dragend', draggableHandleDragEnd, false)
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", sortableInit)
} else {
  sortableInit()
}
