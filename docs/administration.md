Administration
=========

Everything you need to do happens through the filesystem.
The convention Cyber Fire uses is `/srv/moth/state`.

MOTHv5 does "virtual hosting" of instances.
A single server can run as many simultaneous events as your computer can handle.
Each event gets its own directory in `/srv/moth/state/`.
In this document, we'll be using an event named `myevent`,
which will store its state in `/srv/moth/state/myevent`.


Re-initalize
-------------------

    rm initialized

This will reset the following:

* team registrations
* points log

Team tokens stick around, though.


Scores
=======

Pause/resume event
-------------------

    sed -i 's/^enable/#pause/' schedule.txt # pause
    sed -i 's/^#pause/enable/' schedule.txt # resume


Adjusting scores
------------------

    sed -i 's/^enable/#pause/' schedule.txt # pause
    vim awards/*
    sed -i 's/^#pause/enable/' schedule.txt # resume

It's very important to pause the event before mucking around with the points log.
Otherwise, your changes will likely be lost.

There may be more than one file in `awards/`;
if adding an award, you can add it to any file.
If removing an award, remove if everywhere.


Teams
=====

Set a team name / Register a team
------------------

    teamid=e2f8cc14
    echo 'Cool Team Name' > $teamid


List all registered teams
----------------------

    grep . teams/*


Setting up custom team IDs
-------------------

    seq -w 9999 > teamids.txt  # Allow all 4-digit numbers

`teamids.txt` is a list of acceptable team IDs,
one per line.
You can make it anything you want.

New instances will initialize this with some random values.

Remember that team IDs are essentially passwords.


Removing registration checks (Open Registration)
----------------------

    echo '*' > teamids.txt  # Allow every string with length >= 12

⚠️⚠️⚠️ This allows an attack that floods your server with teams. Only do this if you trust your audience. ⚠️⚠️⚠️

The vulnerability can be mitigated to an extent by enabling 
IP Rate Limiting in the HTTPS server front-ending MOTH.
Only the `${interface}/state/teamdata` path needs this.


Disabling team registration
---------------------

    echo > teamids.txt

`teamids.txt` contains a list of team IDs accepted for registration.
An empty list means nobody can register.
You should manually register some teams if you do this!



Dealing with puzzles
===========

Checking on an answer
----------------------

    unzip -p /srv/moth/mothballs/category.mb __MOTH/answers.txt

This will dump every answer for every puzzle.


Installing new categories
-------------------

Just drop a new mothball in the `mothballs' directory.

    cp new-category.mb /srv/moth/mothballs


Taking a category offline
-------------------------

    rm /srv/moth/mothballs/old-category.mb

Removing a category won't remove points that have been scored in it!
