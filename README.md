[![Go report card](https://goreportcard.com/badge/github.com/dirtbags/moth)](https://goreportcard.com/report/github.com/dirtbags/moth)

Quick Start
===========

```sh
# Build MOTH
git clone https://git.woozle.org/neale/moth.git
cd moth
go build -o . ./...
mkdir -p state/demo mothballs

# Create a mothball (puzzle category)
zip -r mothballs/cat0.mb cmd/mothd/testdata/categories/cat0

# Install a web client
git clone https://git.woozle.org/neale/luna.git

# Run it!
./mothd -web luna
```


About
=====================

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

* [Getting Started](docs/getting-started.md): This guide will get you started with a production server.
* [Administration](docs/administration.md): How to set hours, and change setup.


Contributing to MOTH
==================

Please read our [contributing guide](docs/CONTRIBUTING.md).
