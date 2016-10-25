Dirtbags Monarch Of The Hill Server
=====================

This is a set of thingies to run our Monarch-Of-The-Hill contest,
which in the past has been called
"Tracer FIRE",
"Project 2",
"HACK",
"Queen Of The Hill",
"Cyber Spark",
and "Cyber Fire".

Information about these events is at
http://dirtbags.net/contest/

This software serves up puzzles in a manner similar to Jeopardy.
It also tracks scores,
and comes with a JavaScript-based scoreboard to display team rankings.


How everything works
---------------------------

This section wound up being pretty long.
Please check out [the overview](docs/overview.md)
for details.


Getting Started Developing
-------------------------------

    $ git clone $your_puzzles_repo puzzles
    $ python3 tools/devel-server.py

Then point a web browser at http://localhost:8080/
and start hacking on things in your `puzzles` directory.

More on how the devel sever works in
[the devel server documentation](docs/devel-server.md)


Running A Production Server
====================

XXX: Update this

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
