Dirtbags Monarch Of The Hill Server
=====================

This is a set of thingies to run our Monarch-Of-The-Hill contest,
which in the past has been called
"Tracer FIRE",
"Project 2",
"HACK",
"Queen Of The Hill",
and "Cyber FIRE".

Information about these events is at
http://dirtbags.net/contest/

This software serves up puzzles in a manner similar to Jeopardy.
It also track scores,
and comes with a JavaScript-based scoreboard to display team rankings.


How everything works
---------------------------

This section wound up being pretty long.
Please check out [the overview](doc/overview.md)
for details.


Dependencies
--------------------
If you're using Ubuntu 14.04 LTS, you should have everything you need except 
[LUA](http://lua.org).

	$ sudo apt-get install lua5.2 -y

You'll also have to figure out a way to serve up CGI. Here's one way to do it 
on Ubuntu 14.04 LTS with [lighttpd](https://lighttpd.net) where the contest 
is at `/opt/moth/mycontest` accessible by user group `moth`, and the web 
server is run as user `www-data`:

First, make sure your web server can access the contest files:

	$ sudo usermod -a -G moth www-data

Next, install lighttpd and backup the configuration:

	$ sudo apt-get install lighttpd
	$ cp /etc/lighttpd/lighttpd.conf /etc/lighttpd/lighttpd.conf.orig

Add an entry for your contest to your `/etc/hosts`:

	127.0.0.1 mycontest

Add a virtual host entry to `/etc/lighttpd/lighttpd.conf`:

	$HTTP["host"] == "mycontest" {
		server.document-root = "/opt/moth/mycontest/www"
		cgi.assign = ( ".cgi" => "/usr/bin/lua" )
	}

Finally, restart your server:

	$ sudo service lighttpd restart
	* Stopping web server lighttpd  [ OK ] 
	* Starting web server lighttpd  [ OK ]

How to set it up
--------------------

It's made to be virtualized,
so you can run multiple contests at once if you want.
If you were to want to run it out of `/opt/moth`,
do the following:

	$ mkdir -p /opt/moth/mycontest
	$ ./install /opt/moth/mycontest
	$ cp mothd /opt/moth
	
Yay, you've got it set up.


Installing Puzzle Categories
------------------------------------

Puzzle categories are distributed in a different way than the server.
After setting up (see above), just run
	$ /opt/moth/mycontest/bin/install-category /path/to/my/category
	

Running It
-------------

Get your web server to serve up files from
`/opt/moth/mycontest/www`.

Then run `/opt/moth/mothd`.


Permissions
----------------

It's up to you not to be a bonehead about permissions.

Install sets it so the web user on your system can write to the files it needs to,
but if you're using Apache,
it plays games with user IDs when running CGI.
You're going to have to figure out how to configure your preferred web server.
