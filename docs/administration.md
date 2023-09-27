Administration
=========

Everything you need to do happens through the filesystem.
Usually, in `/srv/moth/state`.

The server doesn't cache anything in memory,
so the `state` directory always contains the current state.


Backing up current state
---------------------------

    tar czf backup.tar.gz /srv/moth/state  # Full backup
    curl http://localhost:8080/state > state.json  # Pull anonymized event log and team names (scoreboard)



Scheduling an automatic pause and resume
-----------------------------------

    printf '-'; date --rfc-3339=s -d '10:00 PM' >> /srv/moth/state/hours.txt  # Schedule suspend at 10:00 PM
    printf '+'; date --rfc-3339=s -d '08:00 tomorrow' >> /srv/moth/state/hours.txt # Schedule resume at 08:00 tomorrow

You might prefer to open `/srv/moth/state/hours.txt` in a text editor.
I do.


Re-initalize
-------------------

    rm /srv/moth/state/initialized

This will reset the following:

* team registrations
* points log

Team tokens stick around, though.


Scores
=======

Pausing/resuming scoring
-------------------

    echo '-###' >> /srv/moth/state/hours.txt # Suspend scoring
    sed -i '/###/d' /srv/moth/state/hours.txt # Resume scoring

When scoring is paused,
participants can still submit answers,
and the system will tell them whether the answer is correct.
As soon as you unpause,
all correctly-submitted answers will be scored.


Adjusting scores
------------------

    echo '-###' >> /srv/moth/state/hours.txt # Suspend scoring
    nano /srv/moth/state/points.log  # Replace nano with your preferred editor
    sed -i '/###/d' /srv/moth/state/hours.txt # Resume scoring

We don't warn participants before we do this:
any points scored while scoring is suspended are queued up and processed as soon as scoreing is resumed.

It's very important to suspend scoring before mucking around with the points log.
The maintenance loop assumes it is the only thing writing to this file,
and any edits you make will remove points scored while you were editing.


Teams
=====

Changing a team name
----------------------

    grep . /srv/moth/state/teams/*  # Show all team IDs and names
    echo 'exciting new team name' > /srv/moth/state/teams/$teamid

Please remember, you have to replace `$teamid` with the actual team ID that you want to edit.


Setting up custom team IDs
-------------------

    echo > /srv/moth/state/teamids.txt  # Teams must be registered manually
    seq 9999 > /srv/moth/state/teamids.txt  # Allow all 4-digit numbers

`teamids.txt` is a list of acceptable team IDs,
one per line.
You can make it anything you want.

New instances will initialize this with some hex values.

Remember that team IDs are essentially passwords.


Disabling team registration
---------------------

`teamids.txt` contains a list of team IDs accepted for registration.
If you don't want teams to self-register,
zero out the list:

    true > /srv/moth/state/teamids.txt


Manually registering a team
------------------

    teamid=e2f8cc14
    echo "Cool Team Name" > /srv/moth/state/teams/$teamid


Dealing with puzzles
===========

Checking on an answer
----------------------

Mothballs are just zip files.
If you need to check something about a running category,
just unzip the mothball for that category.

    mkdir /tmp/category
    cd /tmp/category
    unzip /srv/moth/mothballs/category.zip
    cat answers.txt  # Show all valid answers for all puzzles. Watch your shoulder!


Installing new categories
-------------------

Just drop a new mothball in the `mothballs' directory.

    cp new-category.mb /srv/moth/mothballs


Taking a category offline
-------------------------

    rm /srv/moth/mothballs/old-category.mb

Removing a category won't remove points that have been scored in it!
