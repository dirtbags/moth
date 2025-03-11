MOTH Client API
===========

MOTH only has four user API endpoints:

* Retrieve current state
* Upload public team blob
* Retrieve puzzle data
* Submit answer

As a deployment convenience,
the MOTH server will also serve static files.

Authentication
============

MOTH uses HTTP Basic authentication.
Either the password or username may contain the team ID:
password is a better choice, 
since Basic authentication does not allow `:` in usernames.

Clients may use any authenticated API call to "log in".
A 401 return code indicates the team ID is not valid.


State
====

State Object
--------

The state object represents the current state of the server.

State can be obtained anonymously.

If authenticated,
the requesting team's anonymized team ID will be `self` in the returned JSON object.
Clients may use this to display the logged-in team differently on any scoreboards,
or to show teams which puzzles they have already answered.


### `GET` `/{instance}/state/current.json` - Retrieve current state

#### Parameters

| name | type | data type | description |
| ---- | ---- | ---- | ---- |
| `instance` | required | string | MOTH instance |

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Current state in payload (text/plain) |
| 401 | Authentication is invalid (bad team ID) |


#### Example curl

```sh
curl --user teamid:_http://localhost:8080/test1/state/current.json
```

Team Data
--------

Up to 4096 bytes of arbitrary data per team is stored on the server.
This data is public,
but can only be written once.

Until team data exists for a team,
that team ID is not considered valid for any authenticated actions.

Clients may use these blobs however they like:
team name and team color,
or just team name.
In the future,
this document may provide a suggested standard for JSON fields.

MOTHd places no restrictions on the format of this data:
it is stored verbatim and served verbatim.
Clients should be prepared to gracefully handle data that does not look correct:
for instance, a client storing JSON should be able to handle non-JSON data without crashing.


### `PUT` `/{instance}/state/teamdata` - Write team data and register team

This can only happen once per team ID:
further changes must be performed by a MOTH administrator.

The server will return an updated state object,
with the newly-registered team included.

Clients may choose to call this endpoint from a
"register a new new team"
dialog.
Alternately, a client can present login and team registration as the same action,
treating a 409 response as successful login,
optionally with a message that the team has already been created.

#### Parameters

| name | type | data type | description |
| ---- | ---- | ---- | ---- |
| `instance` | required | string | MOTH instance |
| body | required | `application/octet-stream` | public data |

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Team ID is valid, and team name was recorded |
| 401 | Team ID is invalid, or no authentication was sent |
| 409 | Team ID is valid, but team name was previously set and cannot be changed |

#### Example curl

```sh
curl --user teamid:_ -X PUT -H 'Content-type: application/octet-stream' --data '{"name": "bloopers"}' http://localhost:8080/test1/state/public
```


Puzzles
======


Content
--------------

Categories and Puzzles consist of static content,
with the exception of the `answer` file in each puzzle.
Requests must be authenticated with a valid team ID.


### `GET` `/{instance}/puzzles/{category}/{path...}` - Retrieve File

This retrieves both category files and puzzle files,

* `index.html` - Category file
* `1/index.html` - Puzzle file
* `1/salad.jpg` - Puzzle file

See [Puzzle Format](puzzle-format.md) for details on metadata.

#### Parameters

| name | type | data type | description |
| ---- | ---- | ---- | ---- |
| `instance` | required | string | MOTH instance |
| `category` | required | string | Category name |
| `filename` | required | string | File name |

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Content follows |
| 401 | Authentication is required |
| 404 | Category or Filename not found |


#### Example curl

```sh
curl --user teamid:_ http://localhost:8080/test1/puzzles/counting/200/index.html
```



Answer Submission
------------------

### `PUT` `/{instance}/puzzles/{category}/{puzzle}/answer` - Submit Answer

#### Parameters

| name | type | data type | description |
| ---- | ---- | ---- | ---- |
| `instance` | required | string | MOTH instance |
| `category` | required | string | Category name |
| `filename` | required | string | File name |
| body | required | `application/octet-stream` | Answer

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Answer is correct, and award has been recorded |
| 202 | Answer is correct, but award has already been recorded for this puzzle |
| 404 | Puzzle does not exist or is not unlocked |
| 409 | Answer is incorrect |
| 401 | Authentication is invalid (bad team ID) |
| 411 | Content-Length header was not provided |
| 413 | Content-Length value is too large |

#### Example curl

```sh
curl --user teamid:_ -X PUT -H 'Content-type: application/octet-stream' --data "flubber" http://localhost:8080/test1/puzzles/counting/200/answer
```


Static content
======

Non-API URLs are served as static content,
just like a normal web server.


### `GET` `/{path...}` - Retrieve Resource

Directory indexing is enabled:
if a path ends with `/` and no `index.html` is found in the requested directory,
the server provides a directory listing.

#### Parameters

| name | type | data type | description |
| ---- | ---- | ---- | ---- |
| `path` | required | string | Path to resource |

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Content follows |
| 404 | Path not found |


#### Example curl

```sh
curl http://localhost:8080/
```
