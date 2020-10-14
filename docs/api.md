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
        "Devel": true/false
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
