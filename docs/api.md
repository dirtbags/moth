MOTH API
=======

Data encoding
-----

MOTH runs as an HTTP service,
accepting standard HTTP `GET` and `POST`.

Parameters may be encoded with standard `GET` query parameters
(like `GET /endpoint?a=1&b=2`),
or with `POST` as `application/x-www-form-encoded` data.


Endpoints
--------

### `/state`

Returns the current MOTH event state as a JSON object.

#### Parameters
* `id`: team ID (optional)

#### Return

```json
{
    "Config": {
        "Devel": false # true means this is a development server
    },
    "Messages: "HTML to be rendered as broadcast messages",
    "TeamNames": {
        "self": "Requesting team name", # Only if regestered team id is a provided
        "0": "Team 1 Name",
        "1": "Team 2 Name",
        ...
    },
    "PointsLog": {
        [1602679698, "0", "category", 1], # epochTime, teamID, category, points
        ...
    },
    "Puzzles": {
        "category": [1, 2, 3, 6], # list of unlocked puzzles for category
        ...
    }
}
```

#### Example HTTP transaction

##### Request

```
GET /state HTTP/1.0

```

##### Response

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

### `/register`

Registers a name to a team ID.

This is only required once per team,
but user interfaces may find it less confusing to users
to present a "login" page.
For this reason "this team is already registered"
does not return an error.

#### Parameters
* `id`: team ID
* `name`: team name

#### Return

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

#### Example HTTP transaction

##### Request

```
POST /register HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 26

id=b387ca98&name=dirtbags
```

##### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length=86

{"status":"success","data":{"short":"registered","description":"Team ID registered"}}
```


### `/answer`

Submits an answer for points.

If the answer is wrong, no points are awarded 😉

#### Parameters
* `id`: team ID
* `category`: along with `points`, uniquely identifies a puzzle
* `points`: along with `category`, uniquely identifies a puzzle

#### Return

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

#### Example HTTP transaction

##### Request

```
POST /answer HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 62

id=b387ca98&category=sequence&points=2&answer=achilles+turnip
```

##### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length=83

{"status":"fail","data":{"short":"not accepted","description":"Incorrect answer"}}
```


### `/content/{category}/{points}/{filename}`

Retrieves static content associated with a puzzle.

Every puzzle provides `puzzle.json`,
a JSON object containing
information about the puzzle such as the body 
and list of attached files. 

Parameters are all in the URL for this endpoint,
so `curl` and `wget` can be used.

#### Parameters
* `{category}` (in URL): along with `{points}`, uniquely identifies a puzzle
* `{points}` (in URL): along with `{category}`, uniquely identifies a puzzle
* `{filename}` (in URL): filename to retrieve

#### Return

Raw file octets,
with a (hopefully) suitable
`Content-type` HTTP header field.

#### Example HTTP transaction

##### Request

```
GET /content/sequence/1/puzzle.json HTTP/1.0

```

##### Repsonse

```
HTTP/1.0 200 OK
Content-Type: application/json
Content-Length: 397

{"Pre":{"Authors":["neale"],"Attachments":[],"Scripts":[],"Body":"\u003cp\u003e1 2 3 4 5 ⬜\u003c/p\u003e\n","AnswerPattern":"","AnswerHashes":["e7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683"]},"Post":{"Objective":"","Success":{"Acceptable":"","Mastery":""},"KSAs":null},"Debug":{"Log":[],"Errors":[],"Hints":[],"Summary":"Simple introduction to how this works"},"Answers":[]}
```


`puzzle.json`
-------

The special file `puzzle.json` describes a puzzle. It is a JSON object with the following fields:

```json
{
  "Pre": { # Things which appear before the puzzle is solved
    "Authors": ["Neale Pickett"], # List of puzzle authors, usually rendered as a footnote
    "Attachments": ["tiger.jpg"],  # List of files attached to the puzzle
    "Scripts": [],  # List of scripts which should be included in the HTML render of the puzzle
    "Body": "<p>Can you find the hidden text?</p><p><img src=\"tiger.jpg\" alt=\"Grr\" /></p>\n", # HTML puzzle body
    "AnswerPattern": "", # Regular expression to include in HTML input tag for validation
    "AnswerHashes": [ # List of SHA265 hashes of correct answers, for client-side answer checking
      "f91b1fe875cdf9e969e5bccd3e259adec5a987dcafcbc9ca8da62e341a7f29c6"
    ]
  },
  "Post": { # Things reveal after the puzzle is solved
    "Objective": "Learn to examine images for hidden text", # Learning objective
    "Success": { # Measures of learning success
      "Acceptable": "Visually examine image to find hidden text",
      "Mastery": "Visually examine image to find hidden text"
    },
    "KSAs": null # Knowledge, Skills, and Abilities covered by this puzzle
  },
  "Debug": { # Debugging output used in development: all fields are emptied when making mothballs
    "Log": [], # Debug message log
    "Errors": [], # Errors encountered generating this puzzzle
    "Hints": [ # Hints for instructional assistants to provide to participants
        "Zoom in to the image and examine all sections carefully"
    ], 
    "Summary": "text in image" # Summary of this puzzle, to help identify it in an overview of puzzles
  },
  "Answers": ["sandwich"] # List of answers: empty in production
}
```
