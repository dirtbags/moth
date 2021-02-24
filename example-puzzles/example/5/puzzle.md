---
authors: 
  - neale
scripts:
  - filename: helpers.js
  - filename: draggable.js
answers:
  - helper
debug:
  summary: Using JavaScript Input Helpers
---
MOTH only takes static answers:
you can't, for instance, write code to check answer correctness.
But you can provide as many correct answers as you like in a single puzzle.

This page has an associated `helpers.js` script
you can include to assist with input formatting,
so people aren't confused about how to enter an answer.

You could also write your own JavaScript to validate things.

This is just a demonstration page.
You will probably only want one of these in a page,
to avoid confusing people.

### RFC3339 Timestamp
<div class="answer" data-join="">
  <input type="date">
  <input type="hidden" value="T">
  <input type="time" step="1">
  <input type="hidden" value="Z">
</div>

### All lower-case letters
<input class="answer lower">

### Multiple concatenated values
<div class="answer lower">
  <input type="color">
  <input type="number">
  <input type="range" min="0" max="127">
  <input>
</div>

### Free input, sorted, concatenated values
<ul class="answer lower sort">
  <li><input></li>
  <li><button class="expand" title="Add another input">➕</button><l/i>
</ul>

### User-draggable values
<ul class="answer">
  <li draggable="true"><input value="First" readonly></li>
  <li draggable="true"><input value="Third" readonly></li>
  <li draggable="true"><input value="Second" readonly></li>
</ul>

### Select from an ordered list of options
<ul class="answer">
  <li><input type="checkbox" value="horn">Horns</li>
  <li><input type="checkbox" value="hoof">Hooves</li>
  <li><input type="checkbox" value="antler">Antlers</li>
</ul>

### Substring matches
#### Any substring
<input class="answer substring">

#### Only if at the beginning
<input class="answer substring anchor-beg">

#### Only if at the end
<input class="answer substring anchor-end">
