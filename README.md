Dirtbags Monarch Of The Hill Server
=====================

![Build badge](https://github.com/dirtbags/moth/workflows/Build/Test/Push/badge.svg)
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


Run in demonstration mode
===========

    docker run --rm -it -p 8080:8080 dirtbags/moth-devel

Then open http://localhost:8080/ and check out the example puzzles.


Documentation
==========

* [Development](docs/development.md): The development server lets you create and test categories, and compile mothballs.
* [Getting Started](docs/getting-started.md): This guide will get you started with a production server.
* [Administration](docs/administration.md): How to set hours, and change setup.



Contributing to MOTH
==================

Please read our [contributing guide](docs/CONTRIBUTING.md).
