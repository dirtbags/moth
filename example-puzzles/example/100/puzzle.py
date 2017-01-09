#!/usr/bin/python3

import io

def make(puzzle):
    puzzle.author = 'neale'
    puzzle.summary = 'crazy stuff you can do with puzzle generation'

    puzzle.body.write("## Crazy Things You Can Do With Puzzle Generation\n")
    puzzle.body.write("\n")
    puzzle.body.write("The source to this puzzle has some advanced examples of stuff you can do in Python.\n")
    puzzle.body.write("\n")

    # You can use any file-like object; even your own class that generates output.
    f = io.BytesIO("This is some text! Isn't that fantastic?".encode('utf-8'))
    puzzle.add_stream(f)

    # We have debug logging
    puzzle.log("You don't have to disable puzzle.log calls to move to production; the debug log is just ignored at build-time.")
    puzzle.log("HTML is <i>escaped</i>, so you don't have to worry about that!")

    puzzle.answers.append('coffee')
    answer = puzzle.make_answer()        # Generates a random answer, appending it to puzzle.answers too
    puzzle.log("Answers: {}".format(puzzle.answers))

