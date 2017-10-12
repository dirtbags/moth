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


How Scores are Calculated
-------------------------

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


Requirements
-------------

MOTH was written to run on a wide range of Linux systems.
We are very careful not to require exotic extensions:
you can run MOTH equally well on OpenWRT and Ubuntu Server.
It might even run on BSD: if you've tried this, please email us!

Its architecture also limits permissions,
to make it easier to lock things down very tight.
Since it writes to the filesystem slowly and atomically,
it can be run from a USB flash drive formatted with VFAT.


On the server, it requires:

* Bourne shell (POSIX 1003.2: BASH is okay but not required)
* Awk (POSIX 1003.2: gawk is okay but not required)
* Lua 5.1


On the client, it requires:

* A modern web browser with JavaScript
* Categories might add other requirements (like domain-specific tools to solve the puzzles)


Filesystem Layout
=================

The system is set up to make it simple to run one or more contests on a single machine.

I like to use `/srv/moth` as the base directory for all instances.
So if I were running an instance called "hack",
the instance directory would be `/srv/moth/hack`.

There are five entries in each instance directory, described in detail below:

    /srv/moth/hack                 # (r-x) Instance directory
    /srv/moth/hack/assigned.txt    # (r--) List of assigned team tokens
    /srv/moth/hack/bin/            # (r-x) Per-instance binaries
    /srv/moth/hack/categories/     # (r-x) Installed categories
    /srv/moth/hack/state/          # (rwx) Contest state
    /srv/moth/hack/www/            # (r-x) Web server documentroot



`state/assigned.txt`
----------------

This is just a list of tokens that have been assigned.
One token per line, and tokens can be anything you want.

For my middle school events, I make tokens all possible 4-digit numbers,
and tell kids to use any number they want: it makes it quicker to start.
For more advanced events,
this doesn't work as well because people start guessing other teams' numbers to confuse each other.
So I use hex representations of random 32-bit ints.
But you could use anything you want in here (for specifics on allowed characters, read the registration CGI).

The registration CGI checks this list to see if a token has already assigned to a team name.
Teams enter points by token,
which lets them use any text they want for a team name.
Since we don't read their team name anywhere else than the registration and scoreboard generator,
it allows some assumptions about what kind of strings tokens can be,
resulting in simpler code.


`categories/`
--------------

`categories/` contains read-only category packages.
Within each subdirectory there is:

* `map.txt` mapping point values to directory names
* `answers.txt` a list of answers for each point value
* `salt` used to generate directory names (so people can't guess them to skip ahead)
* `summary.txt` a compliation of `00summary.txt` files for puzzles, to give you a quick reference point when someone says "I need help on js 40".
* `puzzles` is all the HTML that needs to be served up for the category


`bin/`
------

Contains all the binaries you'll need to run an event.
These are probably just copies from the `base` package (where this README lives).
They're copied over in case you need to hack on them during an event.

`bin/once` is of particular interest:
it gets run periodically to do everything, including:

* Gather points from `points.new` and append them to the points log.
* Generate a new `puzzles.html` listing all open puzzles.
* Generate a new `points.json` for the scoreboard

### Pausing `once`

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


`www/`
-----------

HTML root for an event.
It is possible to make this read-only,
after you've set up your packages.
You will need to symlink a few things into the `state` directory, though.


`state/`
---------

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
