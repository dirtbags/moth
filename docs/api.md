MOTHv3 API
==========

MOTH, by design, uses a small number of API endpoints.

Whenever possible,
we decided to push complexity into the client,
keeping the server as simple as we could make it.
After all,
this is a hacking contest.
If a participant finds a vulnerability in code running on their own machine,
the people running the server don't care.

Specification
=============

You make requests as HTTP GET query arguments:

    https://server/path/to/endpoint?var1=val1&var2=val2

The server returns a
[JSend](https://labs.omniti.com/labs/jsend) response:

    {
      status: "success",
      data: "Any JS data type here"
    }


Client State
============

The client (or user interacting with the client) needs to remember only one thing:

* teamId: the team ID used to register

A naive client,
like the one we used from 2009-2018,
can ask the user to type in the team ID for every submission.
This is fine.


Endpoints
=========

RegisterTeam(teamId, teamName)
-------------------------------

Register a team name with a team hash.

Parameters:

* teamId: Team's unique identifier (usually a hex value)
* teamName: Team's human-readable name

On success, no data is returned.
On failure, message contains an English explanation of why.

Example:

    https://server/RegisterTeam?teamId=8b1292ca

    {
      status: "success",
      data: nil
    }


GetPuzzleList()
---------------

Return all currently-open puzzles.

Return data:

* puzzles: dictionary mapping from category to a list of point values.


Example:

    https://server/GetPuzzleList

    {
      status: "success",
      data: {
        "puzzles": {
          "sequence": [1, 2],
          "codebreaking": [10],
        }
      }
    }


GetPuzzle(category, points)
--------------------

Return a puzzle.

Parameters:

* category: name of category to fetch from
* points: point value of the puzzle to fetch

Return data:

* authors: List of puzzle authors
* hashes: list of djbhash values of acceptable answers
* files: dictionary of puzzle-associated filenames and their URLs
* body: HTML body of the puzzle


Example:

    https://server/GetPuzzle?category=sequence&points=1

    {
      status: "success",
      data: {
        "authors": ["neale"],
        "hashes": [177627],
        "files": {
          "happy.png": "https://cdn/assets/0904cf3a437a348bea2c49d56a3087c26a01a63c.png"
        },
        "body": "<pre><code>1 2 3 4 5 _\n</code></pre>\n"
    }


GetPointsLog()
---------------

Return the entire points log, and team names.

Return data:

* teams: mapping from team number (int) to team name
* log: list of (timestamp, team number, category, points)

Note: team number may change between calls.


Example:

    https://server/GetEventsLog

    {
      status: "success",
      data: {
        teams: {
          0: "Zelda",
          1: "Defender"
        },
        log: [
          [1526478368, 0, "sequence", 1],
          [1526478524, 1, "sequence", 1],
          [1526478536, 0, "nocode", 1]
        ]
      }
    }


SubmitAnswer(teamId, category, points, answer)
----------------------

Submit an answer to a puzzle.

Parameters:

* teamId: Team ID (optional: if ommitted, answer is verified but no points are awarded)
* category: category name of puzzle
* points: point value of puzzle
* answer: attempted answer

Example:

    https://server/SubmitAnswer?teamId=8b1292ca&category=sequence&points=1&answer=6

    {
      status: "success",
      data: null
    }

SubmitToken(teamId, token)
---------------------

Submit a token for points

Parameters:

* teamId: Team ID
* token: Token being submitted

Return data:

* category: category for which this token awarded points
* points: number of points awarded


Example:

    https://server/SubmitToken?teamId=8b1292ca&token=wat:30:xylep-radar-nanox

    {
      status: "success",
      data: {
        category: "wat",
        points: 30
      }
    }
