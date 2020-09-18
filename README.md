Dirtbags Monarch Of The Hill Server
=====================

![Build badge](https://github.com/dirtbags/moth/workflows/Mothd%20Docker%20build/badge.svg?branch=master)
![Go report card](https://goreportcard.com/badge/github.com/dirtbags/moth)

Monarch Of The Hill (MOTH) is a puzzle server.
We (the authors) have used it for instructional and contest events called
"Tracer FIRE",
"Project 2",
"HACK",
"Queen Of The Hill",
"Cyber Spark",
"Cyber Fire",
"Cyber Fire Puzzles",
and "Cyber Fire Foundry".

Information about these events is at
http://dirtbags.net/contest/

A few things make MOTH different than other Capture The Flag server projects:

* Once any team opens a puzzle, all teams can work on it (high fives to DC949/Orange County for this idea)
* No penalties for wrong answers
* No time-based point deductions (if you're faster, you get to answer more puzzles)
* No internal notion of ranking or score: it only stores an event log, and scoreboards parse it however they want
* All puzzles must be compiled to static content before it can be served up
* The server does very little: most functionality is in client-side JavaScript

You can read more about why we made these decisions in [philosophy](docs/philosophy.md).


Documentation
==========

* [Development](docs/development.md): The development server lets you create and test categories, and compile mothballs.
* [Getting Started](docs/getting-started.md): This guide will get you started with a production server.
* [Administration](docs/administration.md): How to set hours, and change setup.

Running a Production Server
===========================

    docker run --rm -it -p 8080:8080 -v /path/to/moth/state:/state -v /path/to/moth/mothballs:/mothballs:ro dirtbags/moth

You can be more fine-grained about directories, if you like.
Inside the container, you need the following paths:

* `/state` (rw) Where state is stored. Read [the overview](doc/overview.md) to learn what's what in here.
* `/mothballs` (ro) Mothballs (puzzle bundles) as provided by the development server.
* `/theme` (ro) Overrides for the built-in theme.


Contributing to MOTH
==================

Please read [CONTRIBUTING.md](CONTRIBUTING.md)
