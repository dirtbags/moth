MOTH Client API
===========

MOTH provides a WebDAV interface:
this is described in
[MOTH Client Directory Structure](client-structure.md).

This document explains the WebDAV directory structure
as though it were a REST API.
These endpoints are a subset of the functionality provided,
but should be sufficient for many use cases.

Theme
======

Theme files are served as static content,
just like any standard web server.

### `GET` `/theme/${path}` - Retrieve File


Puzzles
======

Static Files
----------

With the exception of the `answer` file,
puzzle files are served as static content.

The entry point to a puzzle is `index.html`:
see [Puzzle Format](puzzle-format.md) 
for details on its structure.

### `GET` `/puzzles/${category}/${points}/${filename}` - Retrieve File


Answer Submission
------------------

### `GET` `/puzzles/${category}/${points}/answer` - not supported

#### Responses

| http code | meaning |
| ---- | ---- |
| 405 | `GET` method is not supported |


### `POST` `/puzzles/${category}/${points}/answer` - Submit Answer

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Answer is correct, and points are awarded |
| 202 | Answer is correct, but points have already been awarded for this puzzle |
| 409 | Answer is incorrect |
| 401 | Authentication is invalid (bad team ID) |


State
====

Points Log
--------

The points log contains a history of correct answer submission.
Each submission is terminated by a newline (`\n`)
and consists of space-separated fields
of the format:

    ${timestamp} ${team_id} ${category} ${points}

### `GET` `/state/points.log` - Retrieve points log

| http code | meaning |
| ---- | ---- |
| 200 | Points log in payload (text/plain) |
| 401 | Authentication is invalid (bad team ID) |


Team Name
--------

### `GET` `/state/self/name` - Retrieve my team name

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Team ID in payload (text/plain) |
| 401 | Authentication is invalid (bad team ID) |


### `POST` `/state/self/name` - Set my team name

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Team ID is valid, and team name was recorded |
| 202 | Team ID is valid, but team name was previously set and cannot be changed |
| 401 | Authentication is invalid (bad team ID) |


Public Data
--------

Up to 4096 bytes of arbitrary public data per team may be stored on the server.
This data can be viewed by any authenticated team.

There are no restrictions on the content of the data:
clients are free to store whatever they want.

### `GET` `/state/${id}/public.bin` - Retrieve public data

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Data follows (application/octet-stream) |
| 401 | Authentication is invalid (bad team ID) |


### `PUT` `/state/${id}/public.bin` - Upload public data

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Data follows (application/octet-stream) |
| 401 | Authentication is invalid (bad team ID) |


Private Data
--------

Up to 4096 bytes of arbitrary data per team may be stored on the server.
This data is only accessible by an authenticated request,
and is private to the authenticated team.

There are no restrictions on the content of the data:
clients are free to store whatever they want.

### `GET` `/state/self/private.bin` - Retrieve private data

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Data follows (application/octet-stream) |
| 401 | Authentication is invalid (bad team ID) |


### `POST` `/state/self/private.bin` - Upload private data

#### Responses

| http code | meaning |
| ---- | ---- |
| 200 | Data follows (application/octet-stream) |
| 401 | Authentication is invalid (bad team ID) |
