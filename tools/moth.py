#!/usr/bin/python3

import argparse
import contextlib
import glob
import hashlib
import io
import importlib.machinery
import mistune
import os
import random
import string
import tempfile

messageChars = b'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'

def djb2hash(buf):
    h = 5381
    for c in buf:
        h = ((h * 33) + c) & 0xffffffff
    return h

@contextlib.contextmanager
def pushd(newdir):
    curdir = os.getcwd()
    os.chdir(newdir)
    try:
        yield
    finally:
        os.chdir(curdir)

# Get a big list of clean words for our answer file.
ANSWER_WORDS = [w.strip() for w in open(os.path.join(os.path.dirname(__file__),
                                                     'answer_words.txt'))]

class PuzzleFile:
    """A file associated with a puzzle.

    path: The path to the original input file. May be None (when this is created from a file handle
          and there is no original input.
    handle: A File-like object set to read the file from. You should be able to read straight
            from it without having to seek to the beginning of the file.
    name: The name of the output file.
    visible: A boolean indicating whether this file should visible to the user. If False,
             the file is still expected to be accessible, but it's path must be known
             (or figured out) to retrieve it."""

    def __init__(self, stream, name, visible=True):
        self.stream = stream
        self.name = name
        self.visible = visible


class Puzzle:
    def __init__(self, category_seed, points):
        """A MOTH Puzzle.

        :param category_seed: A byte string to use as a seed for random numbers for this puzzle.
                              It is combined with the puzzle points.
        :param points: The point value of the puzzle.
        """

        super().__init__()

        self.points = points
        self.author = None
        self.summary = None
        self.answers = []
        self.files = {}
        self.body = io.StringIO()
        self.logs = []
        self.randseed = category_seed * self.points
        self.rand = random.Random(self.randseed)

    def log(self, msg):
        """Add a new log message to this puzzle."""
        self.logs.append(msg)

    def read_stream(self, stream):
        header = True
        for line in stream:
            if header:
                line = line.strip()
                if not line:
                    header = False
                    continue
                key, val = line.split(':', 1)
                key = key.lower()
                if key == 'author':
                    self.author = val
                elif key == 'summary':
                    self.summary = val
                elif key == 'answer':
                    self.answers.append(val)
                elif key == 'file':
                    parts = val.split()
                    name = parts[0]
                    hidden = False
                    stream = open(name, 'rb')
                    try:
                        name = parts[1]
                        hidden = parts[2]
                    except IndexError:
                        pass
                    self.files[name] = PuzzleFile(stream, name, not hidden)
                else:
                    raise ValueError("Unrecognized header field: {}".format(key))
            else:
                self.body.write(line)

    def read_directory(self, path):
        try:
            fn = os.path.join(path, "puzzle.py")
            loader = importlib.machinery.SourceFileLoader('puzzle_mod', fn)
            puzzle_mod = loader.load_module()
        except FileNotFoundError:
            puzzle_mod = None

        if puzzle_mod:
            with pushd(path):
                puzzle_mod.make(self)
        else:
            with open(os.path.join(path, 'puzzle.moth')) as f:
                self.read_stream(f)

    def random_hash(self):
        """Create a file basename (no extension) with our number generator."""
        return ''.join(self.rand.choice(string.ascii_lowercase) for i in range(8))

    def make_temp_file(self, name=None, visible=True):
        """Get a file object for adding dynamically generated data to the puzzle. When you're
        done with this file, flush it, but don't close it.

        :param name: The name of the file for links within the puzzle. If this is None, a name
                     will be generated for you.
        :param visible: Whether or not the file will be visible to the user.
        :return: A file object for writing
        """

        stream = tempfile.TemporaryFile()
        self.add_stream(stream, name, visible)
        return stream

    def add_stream(self, stream, name=None, visible=True):
        if name is None:
            name = self.random_hash()
        self.files[name] = PuzzleFile(stream, name, visible)

    def make_answer(self, word_count, sep=' '):
        """Generate and return a new answer. It's automatically added to the puzzle answer list.
        :param int word_count: The number of words to include in the answer.
        :param str|bytes sep: The word separator.
        :returns: The answer string
        """

        answer = sep.join(self.rand.sample(ANSWER_WORDS, word_count))
        self.answers.append(answer)
        return answer

    def get_body(self):
        return self.body.getvalue()

    def html_body(self):
        """Format and return the markdown for the puzzle body."""
        return mistune.markdown(self.get_body())

    def hashes(self):
        "Return a list of answer hashes"

        return [djbhash(a) for a in self.answers]


class Category:
    def __init__(self, path, seed):
        self.path = path
        self.seed = seed
        self.pointvals = []
        self.catmod = None

        if os.path.exists(os.path.join(path, 'category.py')):
            self.catmod = importlib.machinery.SourceFileLoader(
                'catmod',
                os.path.join(path, 'category.py')).load_module()
            self.pointvals = self.catmod.points[:]
        else:
            for fpath in glob.glob(os.path.join(path, "[0-9]*")):
                pn = os.path.basename(fpath)
                points = int(pn)
                self.pointvals.append(points)

        self.pointvals.sort()

    def puzzle(self, points):
        puzzle = Puzzle(self.seed, points)
        path = os.path.join(self.path, str(points))
        if self.catmod:
            self.catmod.make(points, puzzle)
        else:
            puzzle.read_directory(path)
        return puzzle

    def puzzles(self):
        for points in self.pointvals:
            yield self.puzzle(points)
