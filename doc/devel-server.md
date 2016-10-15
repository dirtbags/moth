Using the MOTH Development Server
======================

To make puzzle development easier,
MOTH comes with a standalone web server written in Python,
which will show you how your puzzles are going to look without making you compile or package anything.

It even works in Windows,
because that is what my career has become.


Starting It Up
-----------------

Just run `devel-server.py` in the top-level MOTH directory.


Installing New Puzzles
-----------------------------

You are meant to have your puzzles checked out into a `puzzles`
directory off the main MOTH directory.
You can do most of your development on this living copy.

In the directory containing `devel-server.py`, you would run something like:

    git clone /path/to/my/puzzles-repository puzzles

or

    ln -s /path/to/my/puzzles-repository puzzles

The development server wants to see category directories under `puzzles`,
like this:

    $ find puzzles -type d
    puzzles/
    puzzles/category1/
    puzzles/category1/10/
    puzzles/category1/20/
    puzzles/category1/30/
    puzzles/category2/
    puzzles/category2/100/
    puzzles/category2/200/
    puzzles/category2/300/

