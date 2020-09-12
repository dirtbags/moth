Developing Content
============================

The development server shows debugging for each puzzle,
and will compile puzzles on the fly.

Use it along with a text editor and shell to create new puzzles and categories.


Set up some example puzzles
---------

If you don't have puzzles of your own to start with,
you can copy the example puzzles that come with the source:

    cp -r /path/to/src/moth/example-puzzles /srv/moth/puzzles


Run the server in development mode
---------------

These recipes run the server in the foreground,
so you can watch the access log and any error messages.


### Podman

    podman run --rm -it -p 8080:8080 -v /srv/moth/puzzles:/puzzles:ro dirtbags/moth -puzzles /puzzles


### Docker

    docker run --rm -it -p 8080:8080 -v /srv/moth/puzzles:/puzzles:ro dirtbags/moth -puzzles /puzzles

### Native

I assume you've built and installed the `moth` command from the source tree.

If you don't know how to build Go packages,
please consider using Podman or Docker.
Building Go software is not a skill related to running MOTH or puzzle events,
unless you plan on hacking on the source code.

    mkdir -p /srv/moth/state
    cp -r /path/to/src/moth/theme /srv/moth/theme
    cd /srv/moth
    moth -puzzles puzzles


Log In
-----

Point a browser to http://localhost:8080/ (or whatever host is running the server).
You will be logged in automatically.


Browse the example puzzles
------------


The example puzzles are written to demonstrate various features of MOTH,
and serve as documentation of the puzzle format.


Make your own puzzle category
-------------------------

    cp -r /srv/moth/puzzles/example /srv/moth/puzzles/my-category


Edit the one point puzzle
--------

    nano /srv/moth/puzzles/my-category/1/puzzle.md

I don't use nano, personally,
but if you're advanced enough to have an opinion about nano,
you're advanced enough to know how to use a different editor.


Read our advice
---------------

The [Writing Puzzles](writing-puzzles.md) document
has some tips on how we approach puzzle writing.
There may be something in here that will help you out!


Stop the server
-------

You can hit Control-C in the terminal where you started the server,
and it will exit.


Mothballs
=======

In the list of puzzle categories and puzzles,
there will be a button to download a mothball.

Once your category is set up the way you like it,
download a mothball for it,
and you're ready to [get started](getting-started.md)
with the production server.
