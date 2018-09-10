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


Getting Started Developing
-------------------------------

You'll want to start out with the Development Server.

More on how the devel sever works in [the devel server documentation](docs/devel-server.md)

To get started quickly, use the Docker image like so `docker run --tm -v /path/to/raw/puzzles/:/puzzles:ro -p 8080:8080 dirtbags/moth-devel` and browse to `http://my_server:8080`


How everything works
---------------------------

This section wound up being pretty long.
Please check out [the overview](docs/overview.md)
for details.


Running A Production Server
====================

Using Docker
-------------------

Dockerfiles and Docker images are provided. 

1. Compile puzzles by running `docker run --rm -v /path/to/puzzle/files:/puzzles:ro -v /path/to/compiled/puzzle/output:/output dirtbags/moth-compile /output /puzzle/<name of category 1> /puzzle/<name of category 2> . . . /puzzle/<name of category N>`. Compiled puzzle zip files are deposited in `/path/to/compiled/puzzle/output`

2. If you want to create pre-specified team hashes, create the file `/path/to/moth/state/assigned.txt` and populate it with valid team hashes, one per line. Otherwise, Moth will auto-generate 100 team hashes

3. Start the Moth server by running `docker run --rm -v /path/to/compiled/puzzle/output/:/moth/puzzles:ro -v /path/to/moth/state/:/moth/state/ -p <listen_port>:80 dirtbags/moth`

4. Use a web browser to visit `http://my_server`

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

`eris` can be obtained at https://github.com/nealey/eris

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
