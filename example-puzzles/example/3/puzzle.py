#!/usr/bin/python3

def make(puzzle):
    puzzle.author = 'neale'
    puzzle.summary = 'dynamic puzzles'
    answer = puzzle.randword()
    puzzle.answers.append(answer)

    puzzle.body.write("To generate a dynamic puzzle, you need to write a Python module.\n")
    puzzle.body.write("\n")
    puzzle.body.write("The passed-in puzzle object provides some handy methods.\n")
    puzzle.body.write("In particular, please use the `puzzle.rand` object to guarantee that rebuilding a category\n")
    puzzle.body.write("won't change puzzles and answers.\n")
    puzzle.body.write("(Participants don't like it when puzzles and answers change.)\n")
    puzzle.body.write("\n")

    number = puzzle.rand.randint(20, 500)
    puzzle.log("One is the loneliest number, but {} is the saddest number.".format(number))

    puzzle.body.write("The answer for this page is {}.\n".format(answer))
