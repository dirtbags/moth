# MOTH's `puzzles` directory

The `puzzles` directory consists of zero or more *categories*:
one per directory.

Each category directory contains one or more *puzzles*:
one per directory.

## Puzzle IDs

Each puzzle directory name is that puzzle's *puzzle ID*.

Puzzle IDs may not contain spaces.

For cross-platform safety,
all Puzzle IDs should be matched by the regular expression
`[-a-zA-Z0-9][-a-zA-Z0-9.]*`

## Examples

Examples can be found in the 
[cmd/mothd/testdata/categories](mothd test data directory);
these are used for unit testing MOTHd.

# Category directories

Category directories consist of a `__MOTH` directory,
containing category metadata,
and one or more puzzle directories.

Category directories may also have top-level files,
such as `index.html`,
which pertain to the entire category.


### `__MOTH`: metadata directory

The `__MOTH` directory contains metadata used by the MOTH server.
The specific files contained are currently:

#### `puzzles.txt`: ordered list of puzzle IDs

This file lists all puzzle IDs in order,
one per line.

### `answers.txt`: puzzle answers

This file lists answers to every puzzle id.
Each line is of the format `$puzzleID $answer`.

Multiple answers may be present for a single puzzle ID:
any answer listed will be accepted as correct.


# Puzzle directories

Each puzzle directory must contain the main puzzle,
and may contain additional attachments.

## `index.html`: puzzle

MOTH puzzles are HTML5 documents,
with optional metadata.
Puzzles may contain stylesheets and scripts,
or any other feature made available by HTML5.

Typically, a puzzle will be rendered in an `<object>` tag
in the MOTH client.
Some clients may copy over scripts, stylesheets,
and embed the puzzle's `<body>` in the page.

### Puzzle metadata

Puzzles may contain metadata,
which can be used by MOTH clients to alter display of puzzles,
or provide additional information in the UI.

Metadata is provided in HTML `<meta>` elements,
with the `name` attribute specifying the metadata name,
and the `content` attribute specifying the metadata content.
Multiple elements with the same `name` are generally permitted.

For example, the following `<meta>` elements
could appear in the `<head>` section of a puzzle's HTML:

```html
<meta name="author" content="Neale Pickett">
<meta name="moth.answerhash" content="87bcc390">
<meta name="moth.answerhash" content="622fcbe8">
<meta name="moth.objective" content="Understand radix 8 (octal)">
```

Metadata names are defined in full in
[MOTH Metadata](metadata.md).



### Images, Attachments, Scripts, and Style Sheets

Related files can be referenced directly in HTML.
Related files *should* be located in the same directory as `index.html`,
but situations may exist where it makes more sense
to locate a file in the parent directory.

Related files are not hidden:
they can be discovered with an http `PROPFIND` method.

For example, assuming `honey.jpg` exists in the same directory
as `index.html`, a standard `<img>` tag will work:

```html
<img src="honey.jpg" 
    alt="A clay jar with the word 'honey' printed on the front." 
    title="Honey jar">
```


### Puzzle Events

As HTML5 documents,
MOTH puzzles can communicate with the MOTH client
using HTML5 events.

#### setAnswer

A MOTH Puzzle may advise the client to fill the answer field with text
by emitting an `setAnswer` custom event.

For example, the following code will advice the client to set the answer field to the string `bloop`:

```javascript
let answerEvent = new CustomEvent(
    "setAnswer", 
    {
        detail: {value: 'bloop'},
        bubbles: true, 
        cancelable: true
    },
)
document.dispatchEvent(answerEvent)
```

MOTH clients *should* listen for such events,
and fill the answer input field with the event's value.
Puzzles *must* provide the user with a copy/paste-able representation of the answer, in the event the event is not handled correctly by the client.


### Example Puzzles

#### Minimally Valid Puzzle

This puzzle provides the absolute minimum required:
a title, and puzzle contents.

```html
<!DOCTYPE html>
<title>Counting</title>
<p>1 2 3 4 5 _</p>
```


#### Puzzle with metadata

Typically, puzzles will provide metadata,
to enable client features such as "possibly correct" validation,
author display, learning objectives


```html
<!DOCTYPE html>
<html>
    <head>
        <title>Counting Sheep</title>
        <meta name="author" content="Neale Pickett">
        <meta name="moth.answerhash" content="089c7244">
        <meta name="moth.answerhash" content="92837b4f">
        <meta name="moth.objective" content="Recognize the difference between a sheep and a wolf">
        <meta name="moth.objective" content="Count to a high number">
        <meta name="moth.success.acceptable" content="Count using fingers">
        <meta name="moth.success.mastery" content="Count using software tools, and provide answer in hexadecimal">
    </head>
    <body>
        <p>🐑🐑🐑🐑🐑🐑🐑🐑🐺🐑🐑</p>
        <p>How many sheep?</p>
    </body>
</html>
```


#### Puzzle with images, scripts, and style

Since they are rendered as HTML documents,
puzzles may include any HTML5 feature.

```html
<!DOCTYPE html>
<html>
    <head>
        <title>Basic Sight Reading</title>
        <meta name="author" content="Neale Pickett">
        <meta name="moth.answerhash" content="baabaa08">
        <meta name="moth.objective" content="Play a tune provided in sheet music">
        <meta name="moth.success.acceptable" content="Play the requested tune with no mistakes">
        <link rel="stylesheet" href="style.css">
        <script src="midi-transcriber.mjs" type="module"></script>
    </head>
    <body>
        <p>
            Using the provided sheet music,
            play "May Had A Little Lamb" on your MIDI keyboard.
            If you make a mistake,
            press the "reset" button and start over.
        </p>
        <p>
            Once you have played from start to finish with no mistakes,
            paste the computed answer into the answer box.
        </p>

        <img src="mary-lamb.png" 
            alt="Sheet music: |EDCD|EEE.|DDD.|EEE.|"
            title="Sheet music for 'Mary Had A Little Lamb">

        <label for="notes">Notes Played</label>
        <output id="notes"></output>
        <button id="reset">Reset</button>

        <label for="answer">Answer</label>
        <output id="answer"></output>
    </body>
</html>
```
