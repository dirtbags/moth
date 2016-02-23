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

Now that your skeleton is set up, you can begin to fill it in.  In each
point-value subdirectory, there can be three special files, and as many
downloadable files as you like, in addition to CGI and any downloadable
but non-listed files you would like.

Special files are:

* index.md: a plain text file formatted with
  [markdown](http://daringfireball.net/projects/markdown/), displayed
  before the list of normal files in the puzzle directory.
* 00answers.txt: a plain text file with acceptable answers, one per line.  Answers
  are matched exactly (ie. they are case-sensitive).
* summary: a single line explaining to contest organizers what's going
  on in this puzzle.

All remaining files, except those with filenames beginning with a comma
(","), are listed on the puzzle page for download.

Any file ending with ".cgi" will be run as CGI.  You can search the web
for how to write a CGI.  Available languages are Python, Lua, and Bourne
Shell.

Let's make our 5-point sandwich question!

    $ cd 5
    $ cat <<EOD >index.mdwn
    > Welcome to the Sandwich category!
    > In this category you will learn how to make a tasty sandwich.
    > The key ingredients in a sandwich are: bread, spread, and filling.
    > When making a sandwich, you need to first put down one slice of bread,
    > then apply any spreads, and finally add filling.  Popular fillings
    > include cheese, sprouts, and cold cuts.  When you are done, apply
    > another slice of bread on top, and optionally tie it together with
    > a fancy toothpick.
    >
    > Now that you know the basics of sandwich-making, it's time for a
    > question!  How many slices of bread are in a sandwich?
    > EOD
    $ cat <<EOD >key
    > 2
    > TWO
    > two
    > EOD
    $ echo "How many slices of bread in a sandwich" > summary
    $

If you wanted to provide a PDF of various sandwiches, this would be the
time to add that too:

    $ cp /tmp/sandwich-types.pdf .
    $

In a real category, you might provide an executable, hard drive image,
or some other kind of blob.

No additional work is needed to have `sandwich-types.pdf` show up as a
download on the puzzle page.



Step 5: Package it up
---------------------

After you've flushed out all your point-value directories, it's time to
wrap it up and send it in.  Clean out any backup or temporary files you
or your editor might have written in the directories, and zip the sucker
up.

    $ cd ../..
    $ zip -r sandwich.zip sandwich/
    $

Now mail the zip file in, and you're all done!
