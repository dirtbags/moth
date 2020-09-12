How to create puzzle categories
===============================

The contest has multiple "puzzle" categories.  Each category contains a
collection of thematically-related puzzles with increasing point
values.  This document will guide you through the process of creating a
new category.  It's up to you to make challenging puzzles, though :)

Since Unix commands are plain text, I'll be using the Unix commands to
illustrate steps.  These are simple commands that should be easy to
translate to a GUI.


Step 1: Establish a progression
-------------------------------

Before you do anything else, you should sit down with a pen and paper,
and plan out how you'd like contestants to progress through your
category.  This contest framework is set up to encourage a linear
progression through puzzles, while still allowing contestants to skip
over things they get stuck on.

The net-re category, for instance, features full tutorial pages with
simple "end of chapter" type questions for point values 1-8.  Point
values 10-99 apply the skills learned in the tutorial against
increasingly challenging problems, point values 100-999 increasingly
approach real-world challenges which use the skills, and point values
1000+ are either culled or inspired by actual net-re tasks performed by
experts in the field.

The crypto category uses the previous answers key as part of the
solution process for each point value.

Ideally, your category will work standalone for novices, while allowing
experts to quickly answer the training questions and progress to real
challenges.  Remember that some events don't have a class portion, and
even the ones that do have students who prefer to spend the contest time
reviewing the exact same problems they did in the class.

Remember, it's easy to make incredibly challenging puzzles, and you will
probably have a lot of ideas about how to do this.  What's harder is to
make simple puzzles that teach.  It can be helpful to imagine a student
with a basic skill set.  Write your first puzzle for this student to
introduce them to the topic and get them thinking about things you
believe are important.  Guide that student through your tutorial
puzzles, until they emerge ready to tackle some non-tutorial problems.
As they gain confidence, keep them on their toes with new challenges.
Remember to only introduce one new concept for each puzzle!

Past a certain point, feel free to throw in the killer tricky puzzles
you're just dying to create!



Step 2: Establish point values
------------------------------

Each of your steps needs a point value.  Each point value must be
unique: you may not have two 5-point puzzles.

Point values should roughly reflect how difficult a problem is to solve.
It's not terribly important that a 200-point puzzle be ten times harder
than a 20-point puzzle, but it is crucial that a 25-point puzzle be
roughly as difficult as a 20-point puzzle.  Poorly-weighted puzzles has
been the main reason students lose interest.



Step 3: Set up your puzzle structure
------------------------------------

The best way to get puzzles to me is in a zip file of an entire
directory.  Let's say you are going to create a "sandwich" category.
Your first step will be to make a "sandwich" directory somewhere.

    $ mkdir sandwich
    $ cd sandwich
    $

Within your category directory, create subdirectories for each point
value puzzle.  In the "sandwich" category we have only 5, 10, and
100-point puzzles.

    $ mkdir 5 10 100
    $


Step 4: Write puzzles
---------------------

Now that your skeleton is set up, you can begin to fill it in.
Check the `example-puzzles` directory for examples of how to format puzzles,
and how to use the Python Puzzle object for dynamically-generated puzzles.
