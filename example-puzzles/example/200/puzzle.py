import io
import categorylib  # Category-level libraries can be imported here

def make(puzzle):
    import puzzlelib  # puzzle-level libraries can only be imported inside of the make function
    puzzle.authors = ['donaldson']
    puzzle.summary = 'more crazy stuff you can do with puzzle generation using Python libraries'

    puzzle.body.write("## Crazy Things You Can Do With Puzzle Generation (part II)\n")
    puzzle.body.write("\n")
    puzzle.body.write("The source to this puzzle has some more advanced examples of stuff you can do in Python.\n")
    puzzle.body.write("\n")
    puzzle.body.write("1 == %s\n\n" % puzzlelib.getone(),)
    puzzle.body.write("2 == %s\n\n" % categorylib.gettwo(),)

    puzzle.answers.append('tea')
    answer = puzzle.make_answer()        # Generates a random answer, appending it to puzzle.answers too
    puzzle.log("Answers: {}".format(puzzle.answers))

