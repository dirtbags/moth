Overview of MOTH
================

Monarch Of The Hill (MOTH) is a framework for running puzzle-based events for teams.
Each team is assigned a token which they use to identify themselves.

Teams are presented with a number of *categories*,
each containing a sequence of *puzzles* of increasing point value.

When the event starts, only the lowest-point puzzle in each category is available.
As soon as any team enters the correct solution to the puzzle,
the next puzzle is opened up for all teams.

A scoreboard tracks team rankings,
indicating score within each category,
and overall ranking.


State Directory
===============

The state directory is written to by the server to preserve state.
At no point is anything only in memory:
if it's not on the filesystem,
mothd doesn't think it exists.

The state directory is also used to communicate actions to mothd.


`initialized`
-------------

Remove this file to reset the state. This will blow away team assignments and the points log.


`hours.txt`
-------

A list of start and stop hours.
If all the hours are in the future, the event defaults to running.
"Stop" here just pertains to scoreboard updates and puzzle unlocking.
People can still submit answers and their awards are queued up for the next start.


`teamids.txt`
-------------

A list of valid Team IDs, one per line.
It defaults to all 4-digit natural numbers,
but you can put whatever you want in here.


`points.log`
------------

The log of awarded points:

    EpochTime TeamId Category Points

Do not write to this file, unless you have disabled the contest. You will lose points!


`points.tmp`
------------

Drop points logs here.
Filenames can be anything.

When the file is complete and written out,
move it into `points.new`,
where a non-disabled event's maintenance loop will eventually move it into the main log.

`points.new`
------------

Complete points logs should be atomically moved here.
This is to avoid needing locks.
[Read about Maildir](https://en.wikipedia.org/wiki/Maildir)
if you care about the technical reasons we do things this way.


Mothball Directory
==================

Put a mothball in this directory to open that category.
Remove a mothball to disable that category.

Overwriting a mothball with a newer version will be noticed by the server within one maintenance interval
(20 seconds by default).
Be sure to use the same compilation seed in the development server if you compile a new version!

Removing a category does not remove points that have been scored in the category.


Resources Directory
===================


Making it look better
-------------------

`mothd` provides some built-in HTML for rendering a complete contest,
but it's rather bland.
You can override everything by dropping a new file into the `resources` directory:

* `basic.css` is used by the default HTML to pretty things up
* `index.html` is the landing page, which asks to register a team
* `puzzle.html` renders a puzzle from JSON
* `puzzle-list.html` renders the list of active puzzles from JSON
* `scoreboard.html` renders the current scoreboard from JSON
* Any other file in the `resources` directory will be served up, too.

If you don't want to read through the source code, I don't blame you.
Run a `mothd` server and pull the various static resources into your `resources` directory,
and then you can start hacking away at them.


Making it look totally different
---------------------

Every handler can serve its answers up in JSON format,
just add `application/json` to the `Accept` header of your request.

This means you could completely ignore the file structure in the previous section,
and write something like a web app that only loads static resources at startup.


Changing scoring
--------------

Scoring is determined client-side in the scoreboard,
from the points log.
You can hack in whatever algorithm you like,
and provide your own scoreboard(s).

If you do hack in a new algorithm,
please be a dear and email it to us.
We'd love to see it!



How Scores are Calculated by Default
------------------------------------

The per-category score for team `t` is computed as:

* Let `m` be the sum of all points in currently-visible puzzles in this category
* Let `s` be the sum of all points team `t` has won in this category
* Team `t`'s score is `s`/`m`

Therefore, the maximum number of points a team can have in a category is 1.0.
Put another way, a team's per-category score is the percentage of points they've made so far in that category.

The total score for team `t` is the sum of all per-category points.

Therefore, the maximum number of points a team can have in a 7-category event is 7.0.

This point system has proven both easy to explain (if not immediately obvious),
and acceptable by participants.
Because we don't award extra points for quick responses,
teams always feel like they have the possibility to catch up if they are skilled enough.


