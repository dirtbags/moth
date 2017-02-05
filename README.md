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

If you don't have a `puzzles` directory,
you can copy the example puzzles as a starting point:

    $ cp -r example-puzzles puzzles

Then launch the development server:

    $ python3 tools/devel-server.py

Point a web browser at http://localhost:8080/
and start hacking on things in your `puzzles` directory.

More on how the devel sever works in
[the devel server documentation](docs/devel-server.md)


Running A Production Server
====================

XXX: Update this

How to install it
--------------------

It's made to be virtualized,
so you can run multiple contests at once if you want.
If you were to want to run it out of `/srv/moth`,
do the following:

    $ mothinst=/srv/moth/mycontest
	$ mkdir -p $mothinst
	$ install.sh $mothinst
	
    Yay, you've got it installed.

How to run a contest
------------------------

`mothd` runs through every contest on your server every few seconds,
and does housekeeping tasks that make the contest "run".
If you stop `mothd`, people can still play the contest,
but their points won't show up on the scoreboard.

A handy side-effect here is that if you need to meddle with the points log,
you can just kill `mothd`,
do you work,
then bring `mothd` back up.

    $ cp src/mothd /srv/moth
    $ /srv/moth/mothd

You're also going to need a web server if you want people to be able to play.


How to run a web server
-----------------------------

Your web server needs to serve up files for you contest out of
`$mothinst/www`.

If you don't want to fuss around with setting up a full-featured web server,
you can use `tcpserver` and `eris`,
which is what we use to run our contests.

`tcpserver` is part of the `uscpi-tcp` package in Ubuntu.
You can also use busybox's `tcpsvd` (my preference, but a PITA on Ubuntu).

`eris` can be obtained at https://woozle.org/neale/g.cgi/net/eris/about/

    $ mothinst=/srv/moth/mycontest
    $ $mothinst/bin/httpd


Installing Puzzle Categories
------------------------------------

Puzzle categories are distributed in a different way than the server.
After setting up (see above), just run

	$ /srv/koth/mycontest/bin/install-category /path/to/my/category
	

Permissions
----------------

It's up to you not to be a bonehead about permissions.

Install sets it so the web user on your system can write to the files it needs to,
but if you're using Apache,
it plays games with user IDs when running CGI.
You're going to have to figure out how to configure your preferred web server.
