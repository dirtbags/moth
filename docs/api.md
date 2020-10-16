Moth APIs
=======

This document covers the following interfaces:

* HTTP Endpoints: what the Moth client sends the Moth server
* Puzzle executable: how the transpiler communicates with executables that provide puzzles
* Category executable: how the transpiler communicates with executables that provide categories
* Provider executable: how Moth communicates with things that provide puzzles (like the transpiler)

The Puzzle, Category, and Provider executalbes are all very closely related, since each is a subset of the next.

----

Here's a bad diagram of how this all fits together. I don't know if this is going to help at all. Please submit a merge request with something better.

                 HTTP    provider API           mothball API
              🡗           🡗                           🡗
    client - mothd - mothball provider - category1.mb

                         - custom provider
                                                        category API
                                                      🡗
                         - internal transpiler - category2/mkcategory

                                                      - category3/1/puzzle.md
                                                      - category3/2/mkpuzzle
                                                       🡔 
                                                         puzzle API

                                                                      

# HTTP Endpoints

The Moth server accepts
standard HTTP `GET` and `POST`.

Parameters may be encoded with standard `GET` query parameters
(like `GET /endpoint?a=1&b=2`),
or with `POST` as `application/x-www-form-encoded` data.

## `/state`

Returns the current Moth event state as a JSON object.

### Parameters
* `id`: team ID (optional)

### Return

```js
{
    "Config": {
        "Devel": false // true means this is a development server
    },
    "Messages: "HTML to be rendered as broadcast messages",
    "TeamNames": {
        "self": "Requesting team name", // Only if regestered team id is a provided
        "0": "Team 1 Name",
        "1": "Team 2 Name"
        // ...
    },
    "PointsLog": [
        [1602679698, "0", "category", 1] // epochTime, teamID, category, points
        // ...
    ],
    "Puzzles": {
        "category": [1, 2, 3, 6] // list of unlocked puzzles for category
        // ...
    }
}
```

### Example HTTP transaction

#### Request

```
GET /state HTTP/1.0

```

#### Response

This response has been reflowed for readability:
an actual on-wire response would not have newlines or indentation.

```
HTTP/1.0 200 OK
Content-Type: application/json

{"Config":
  {"Devel":false},
  "Messages":"<p>Welcome to the event!</p><p>Event ends at 19:00!</p>",
  "TeamNames":{
      "0":"Mike and Jack",
      "12":"Team 2",
      "4":"Team 8"
    },
    "PointsLog":[
        [1602702696,"0","nocode",1],
        [1602702705,"0","sequence",1],
        [1602702787,"0","nocode",2],
        [1602702831,"0","sequence",2],
        [1602702839,"4","nocode",3],
        [1602702896,"0","sequence",8],
        [1602702900,"4","nocode",4],
        [1602702913,"0","sequence",16]
    ],
    "Puzzles":{
        "indy":[12],
        "nocode":[1,2,3,4,10],
        "sequence":[1,2,8,16,19],
        "steg":[1]
    }
}
```

## `/register`

Registers a name to a team ID.

This is only required once per team,
but user interfaces may find it less confusing to users
to present a "login" page.
For this reason "this team is already registered"
does not return an error.

### Parameters
* `id`: team ID
* `name`: team name

### Return

An object inspired by [JSend](https://github.com/omniti-labs/jsend):

```json
{
    "status": "success/fail/error",
    "data": {
        "short": "short description",
        "description": "long description"
    }
}
```

### Example HTTP transaction

#### Request

```
POST /register HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 26

id=b387ca98&name=dirtbags
```

#### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length=86

{"status":"success","data":{"short":"registered","description":"Team ID registered"}}
```


## `/answer`

Submits an answer for points.

If the answer is wrong, no points are awarded 😉

### Parameters
* `id`: team ID
* `category`: along with `points`, uniquely identifies a puzzle
* `points`: along with `category`, uniquely identifies a puzzle

### Return

An object inspired by [JSend](https://github.com/omniti-labs/jsend):

```json
{
    "status": "success/fail/error",
    "data": {
        "short": "short description",
        "description": "long description"
    }
}
```

### Example HTTP transaction

#### Request

```
POST /answer HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 62

id=b387ca98&category=sequence&points=2&answer=achilles+turnip
```

#### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length=83

{"status":"fail","data":{"short":"not accepted","description":"Incorrect answer"}}
```

## `/content/{category}/{points}/puzzle.json`

Retrieves the JSON object describing a puzzle.

Parameters are all in the URL for this endpoint,
so `curl` and `wget` can be used.

### Parameters
* `{category}` (in URL): along with `{points}`, uniquely identifies a puzzle
* `{points}` (in URL): along with `{category}`, uniquely identifies a puzzle
* `{filename}` (in URL): filename to retrieve

### Return

JSON object describing a puzzle.

```js
{
  "Pre": { // Things which appear before the puzzle is solved
    "Authors": ["Neale Pickett"], // List of puzzle authors, usually rendered as a footnote
    "Attachments": ["tiger.jpg"],  // List of files attached to the puzzle
    "Scripts": [],  // List of scripts which should be included in the HTML render of the puzzle
    "Body": "<p>Can you find the hidden text?</p><p><img src=\"tiger.jpg\" alt=\"Grr\" /></p>\n", // HTML puzzle body
    "AnswerPattern": "", // Regular expression to include in HTML input tag for validation
    "AnswerHashes": [ // List of SHA265 hashes of correct answers, for client-side answer checking
      "f91b1fe875cdf9e969e5bccd3e259adec5a987dcafcbc9ca8da62e341a7f29c6"
    ]
  },
  "Post": { // Things reveal after the puzzle is solved
    "Objective": "Learn to examine images for hidden text", // Learning objective
    "Success": { // Measures of learning success
      "Acceptable": "Visually examine image to find hidden text",
      "Mastery": "Visually examine image to find hidden text"
    },
    "KSAs": null // Knowledge, Skills, and Abilities covered by this puzzle
  },
  "Debug": { // Debugging output used in development: all fields are emptied when making mothballs
    "Log": [ // Debug message log
      "Input image size: 600x400",
      "Applying gaussian blur",
      "Text width 58, left offset 513",
      "Complete in 0.028s"
    ],
    "Errors": [], // Errors encountered generating this puzzzle
    "Hints": [ // Hints for instructional assistants to provide to participants
        "Zoom in to the image and examine all sections carefully"
    ], 
    "Summary": "text in image" // Summary of this puzzle, to help identify it in an overview of puzzles
  },
  "Answers": ["sandwich"] // List of answers: empty in production
}
```


### Example HTTP transaction

#### Request

```
GET /content/sequence/1/puzzle.json HTTP/1.0

```

#### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length: 397

{"Pre":{"Authors":["neale"],"Attachments":[],"Scripts":[],"Body":"\u003cp\u003e1 2 3 4 5 ⬜\u003c/p\u003e\n","AnswerPattern":"","AnswerHashes":["e7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683"]},"Post":{"Objective":"","Success":{"Acceptable":"","Mastery":""},"KSAs":null},"Debug":{"Log":[],"Errors":[],"Hints":[],"Summary":"Simple introduction to how this works"},"Answers":[]}
```


## `/content/{category}/{points}/{filename}`

Retrieves static content associated with a puzzle.

Parameters are all in the URL for this endpoint,
so `curl` and `wget` can be used.

### Parameters
* `{category}` (in URL): along with `{points}`, uniquely identifies a puzzle
* `{points}` (in URL): along with `{category}`, uniquely identifies a puzzle
* `{filename}` (in URL): filename to retrieve

### Return

Raw file octets,
with a (hopefully) suitable
`Content-type` HTTP header field.

### Example HTTP transaction

#### Request

```
GET /content/sequence/1/attachment.txt HTTP/1.0

```

#### Repsonse

```
HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 98

This is an attachment file! This is just plain text for the example. Many attachments are JPEGs.
```


# Puzzle

A puzzle contains one question and one or more associated answers.
Puzzles are not aware of their point value: this is set by the category they are in.

Puzzle executables must be named `mkpuzzle`.


## `mkpuzzle puzzle`

```
puzzles/category3/1 $ ./mkpuzzle puzzle
{JSON PUZZLE OBJECT}
```


## `mkpuzzle file {filename}`

```
puzzles/category3/1 $ ./mkpuzzle file attachment.txt
This is an attachment file! It's just plain text for this example. Many attachments are JPEGs.
```


## `mkpuzzle answer {answer}`

```
puzzles/category3/1 $ ./mkpuzzle answer "cow goes moo"
{"Correct":false}
```


# Category

Categories are collections of puzzles.
Each puzzle has a unique point value, determined by the category.

Category executables must be called `mkcategory`.

## `mkcategory inventory`

```
puzzles/category2 $ ./mkcategory inventory
{"Inventory": [1, 2, 3, 5, 10, 20, 30, 50, 100]}
```


## `mkcategory puzzle {points}`

```
puzzles/category2 $ ./mkcategory puzzle 1
{JSON PUZZLE OBJECT}
```


## `mkcategory file {points} {filename}`

```
puzzles/category2 $ ./mkcategory file 1 attachment.txt
This is an attachment file's contents!
```


## `mkcategory answer {points} {answer}`

```
puzzles/category2 $ ./mkcategory answer 1 "cow goes moo"
{"Correct":false}
```



# Provider API

This is how Claire gets her dynamic graders.

*Notice: this is not complete in the code base!*
I'm writing here how it *should* work.
If anybody wants this,
please let me know,
and I'll finish the code.

This could ostensibly be expanded to call HTTP servers,
with the four endpoints described here.
If somebody were to want such a thing.

## Inventory

    $ provider inventory
    {
      "category1": [1, 2, 3, 4, 5, 10, 20, 30],
      "category2": [20, 40, 70, 150]
    }

## Puzzle

    $ provider puzzle category1 20
    {JSON PUZZLE OBJECT}

## Attachment

    $ provider file category1 20 attachment.txt
    This is an attachment! Yay!

##  Answer

    $ provider answer category1 20 "cow goes moo"
    {"Correct":true}
