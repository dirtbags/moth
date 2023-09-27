Scoring
=======

MOTH does not carry any notion of who is winning: we consider this a user
interface issue. The server merely provides a timestamped log of point awards.

The bundled scoreboard provides one way to interpret the scores: this is the
main algorithm we use at Cyber Fire events. We use other views of the scoreboard
in other contexts, though! Here are some ideas:


Percentage of Each Category
---------------------

This is implemented in the scoreboard distributed with MOTH, and is how our
primary score calculation at Cyber Fire.

For each category:

* Divide the team's score in this category by the highest score in this category
* Add that to the team's overall score

This means the highest theoretical score in any event is the number of open
categories.

This algorithm means that point values only matter relative to other point
values within that category. A category with 5 total points is worth the same as
a category with 5000 total points, and a 2 point puzzle in the first category is
worth as much as a 2000 point puzzle in the second.

One interesting effect here is that a team solving a previously-unsolved puzzle
will reduce everybody else's ranking in that category, because it increases the
divisor for calculating that category's score.

Cyber Fire used to not display overall score: we would only show each team's
relative ranking per category. We may go back to this at some point!


Category Completion
----------------

Cyber Fire also has a scoreboard called the "class" scoreboard, which lists each
team, and which puzzles they have completed. This provides instructors with a
graphical overview of how people are progressing through content. We can provide
assistance to the general group when we see that a large number of teams are
stuck on a particular puzzle, and we can provide individual assistance if we see
that someone isn't keeping up with the class.


Monarch Of The Hill
----------------

You could also implement a "winner takes all" approach: any team with the
maximum number of points in a category gets 1 point, and all other teams get 0.


Time Bonuses
-----------

If you wanted to provide extra points to whichever team solves a puzzle first,
this is possible with the log. You could either boost a puzzle's point value or
decay it; either by timestamp, or by how many teams had solved it prior.


Bonkers Scoring
-------------

Other zany options exist:

* The first team to solve a puzzle with point value divisible by 7 gets double
  points. 
* [Tokens](tokens.md) with negative point values could be introduced, allowing
  teams to manipulate other teams' scores, if they know the team ID.
