Download All Unlocked Puzzles
========================

We get a lot of requests to "download everything" from an event.
Here's how you could do that:

What You Need
------------

* The URL to your puzzle server. We will call this `$url`.
* Your Team ID. We will call this `$teamid`.
* A way to POST `$teamid` to a URL, and save the result. We will call this procedure "Fetch".
* A way to parse JSON files

Steps
-----

1. Fetch `$url/state`. This is the State object.
2. In the State object, `Puzzles` maps category name to a list of open puzzle point values.
3. For each category (we will call this `$category`):
    1. For each point value:
        1. If the point value is 0, skip it. 0 indicates all puzzles in this category are unlocked.
        2. Fetch `$url/content/$category/$points/index.json`. This is the Puzzle object.
        3. In the Puzzle object, `Body` contains the HTML body of the puzzle.
        4. For each file listed in `Attachments` (we will call this `$attachment`):
            1. Fetch `$url/content/$category/$points/$attachment`.
