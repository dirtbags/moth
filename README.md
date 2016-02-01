Dirtbags King Of The Hill Server
=====================

This is a set of thingies to run our KOTH-style contest.
Contests we've run in the past have been called
"Tracer FIRE" and "Project 2".

It serves up puzzles in a manner similar to Jeopardy.
It also track scores,
and comes with a JavaScript-based scoreboard to display team rankings.


How Everything Works
----------------------------

### `assigned.txt`

This is just a list of tokens that have been assigned.
One token per line, and tokens can be anything you want.

For my middle school events, I make tokens all possible 4-digit numbers,
and tell kids to use any number they want: it makes it quicker to start.
For more advanced events,
this doesn't work as well because people start guessing other teams' numbers to confuse each other.
So I use hex representations of random 32-bit ints.
But you could use anything you want in here (with some restrictions, detailed in the registration CGI).

The registration CGI checks this list to see if a token has already assigned to a team name.
Teams enter points by token,
which lets them use any text they want for a team name.
Since we don't read their team name anywhere else than the registration and scoreboard generator,
it allows some assumptions about what kind of strings tokens can be,
resulting in simpler code.

### `packages/`

`packages/` contains read-only package archives.
Within each subdirectory there is:

* `map.txt` mapping point values to directory names
* `answers.txt` a list of answers for each point value
* `salt` used to generate directory names (so people can't guess them to skip ahead)
* `summary.txt` a compliation of `00summary.txt` files for puzzles, to give you a quick reference point when someone says "I need help on js 40".
* `puzzles` is all the HTML that needs to be served up for the category

### `bin/`

Contains all the binaries you'll need to run an event.
These are probably just copies from the `base` package (where this README lives).
They're copied over in case you need to hack on them during an event.

`bin/once` is of particular interest:
it gets run periodically to do everything, including:

* Gather points from `points.new` and append them to the points log.
* Generate a new `puzzles.html` listing all open puzzles.
* Generate a new `points.json` for the scoreboard

#### Pausing `once`

You can pause everything `bin/once` does by touching a file in the root directory
called `disabled`.
This doesn't stop the game:
it just stops points collection and generation of the files listed above.

This is extremely helpful when, inevitably,
you need to hack the points log,
or do other maintenance tasks.
Most times you don't even need to announce that you're doing anything:
people can keep playing the game and their points keep collecting,
ready to be appended to the log when you're done and you re-enable `once`.


### `www/`

HTML root for an event.
It is possible to make this read-only,
after you've set up your packages.
You will need to symlink a few things into the `state` directory, though.


### `state/`

Where all game state is stored.
This is the only part of the contest directory setup that needs to be writable,
and tarring it up preserves exactly the entire contest.

Notable, it contains the mapping from team hash to name,
and the points log.

`points.log` is replayed by the scoreboard generator to calculate the current score for each team.

New points are written to `points.new`, and picked up by `bin/once` to append to `points.log`.
When `once` is disabled (by touching a file called `disabled` at the top level for a game),
the various points-awarding things can keep writing files into `points.new`,
with no need for locking or "bringing down the game for maintenance".



How to set it up
--------------------

It's made to be virtualized,
so you can run multiple contests at once if you want.
If you were to want to run it out of `/opt/koth`,
do the following:

	$ mkdir -p /opt/koth/mycontest
	$ ./install /opt/koth/mycontest
	$ cp kothd /opt/koth
	
Yay, you've got it set up.


Installing Puzzle Categories
------------------------------------

Puzzle categories are distributed in a different way than the server.
After setting up (see above), just run

	$ /opt/koth/mycontest/bin/install-category /path/to/my/category
	

Running It
-------------

Get your web server to serve up files from
`/opt/koth/mycontest/www`.

Then run `/opt/koth/kothd`.


Permissions
----------------

It's up to you not to be a bonehead about permissions.

Install sets it so the web user on your system can write to the files it needs to,
but if you're using Apache,
it plays games with user IDs when running CGI.
You're going to have to figure out how to configure your preferred web server.
