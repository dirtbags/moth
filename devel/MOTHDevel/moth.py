#!/usr/bin/python3
"""Module containing MOTH puzzle container definitions"""

import contextlib
import glob
import html
import io
import importlib.machinery
import os
import random
import string
import tempfile
import shlex

from . import mistune


def djb2hash(instr):
    """Calculate the DJB2 hash of the input

    :param instr: data to calculate the DJB2 hash digest of
    :return: DJB2 hash digest
    """
    hash_digest = 5381
    for char in instr.encode("utf-8"):
        hash_digest = ((hash_digest * 33) + char) & 0xffffffff
    return hash_digest


@contextlib.contextmanager
def pushd(newdir):
    """Context manager for limiting context to individual puzzles/categories"""
    curdir = os.getcwd()
    os.chdir(newdir)
    try:
        yield
    finally:
        os.chdir(curdir)


def loadmod(name, path):
    """Load a specified puzzle module

    :param name: Name to load the module as
    :param path: Path of the module to load
    """
    abspath = os.path.abspath(path)
    loader = importlib.machinery.SourceFileLoader(name, abspath)
    return loader.load_module()  # pylint: disable=no-value-for-parameter


# Get a big list of clean words for our answer file.
ANSWER_WORDS = [w.strip() for w in open(os.path.join(os.path.dirname(__file__),
                                                     'answer_words.txt'))]


class PuzzleFile:  # pylint: disable=too-few-public-methods

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


class Puzzle:  # pylint: disable=too-many-instance-attributes
    """A MOTH Puzzle.

    :param category_seed: A byte string to use as a seed for random numbers for this puzzle.
                          It is combined with the puzzle points.
    :param points: The point value of the puzzle.
    """
    def __init__(self, category_seed, points):

        self.points = points
        self.summary = None
        self.authors = []
        self.answers = []
        self.scripts = []
        self.pattern = None
        self.hint = None
        self.files = {}
        self.body = io.StringIO()
        self.logs = []
        self.randseed = category_seed * self.points
        self.rand = random.Random(self.randseed)

    def log(self, *vals):
        """Add a new log message to this puzzle."""
        msg = ' '.join(str(v) for v in vals)
        self.logs.append(msg)

    def read_stream(self, stream):  # pylint: disable=too-many-branches
        """Read in a MOTH-formatted puzzle definition

        :param stream: file-like object containing line-separated MOTH definitions
        """
        header = True
        for line in stream:
            if header:
                line = line.strip()
                if not line:
                    header = False
                    continue
                key, val = line.split(':', 1)
                key = key.lower()
                val = val.strip()
                if key == 'author':
                    self.authors.append(val)
                elif key == 'summary':
                    self.summary = val
                elif key == 'answer':
                    self.answers.append(val)
                elif key == 'pattern':
                    self.pattern = val
                elif key == 'hint':
                    self.hint = val
                elif key == 'name':
                    pass
                elif key == 'file':
                    parts = shlex.split(val)
                    name = parts[0]
                    hidden = False
                    stream = open(name, 'rb')

                    try:
                        name = parts[1]
                        hidden = (parts[2].lower() == "hidden")
                    except IndexError:
                        pass

                    self.files[name] = PuzzleFile(stream, name, not hidden)
                elif key == 'script':
                    stream = open(val, 'rb')
                    # Make sure this shows up in the header block of the HTML output.
                    self.files[val] = PuzzleFile(stream, val, visible=False)
                    self.scripts.append(val)
                else:
                    raise ValueError("Unrecognized header field: {}".format(key))
            else:
                self.body.write(line)

    def read_directory(self, path):
        """Read a puzzle definition out of a directory.

        :param path: Path to the directory containing a puzzle.py or puzzle.moth file
        """
        try:
            puzzle_mod = loadmod("puzzle", os.path.join(path, "puzzle.py"))
        except FileNotFoundError:
            puzzle_mod = None

        with pushd(path):
            if puzzle_mod:
                puzzle_mod.make(self)
            else:
                with open('puzzle.moth') as puzzle_file:
                    self.read_stream(puzzle_file)

    def random_hash(self):
        """Create a file basename (no extension) with our number generator.

        :return: An 8-character random name
        """
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
        """Add a file-like object to a puzzle

        :param stream: file-like object to write to a puzzle file
        :param name: Name to assign the data. If this is None, a name will be generated for you
        :param visible: Whether or not the file will be visible to the user.
        """
        if name is None:
            name = self.random_hash()

        self.files[name] = PuzzleFile(stream, name, visible)

    def add_file(self, filename, visible=True):
        """Add a file to a puzzle

        :param filename: Name of the file to add to the puzzle
        :param visible: Whether or not the file will be visible to the user.
        """
        file_handle = open(filename, 'rb')
        name = os.path.basename(filename)
        self.add_stream(file_handle, name=name, visible=visible)

    def randword(self):
        """Return a randomly-chosen word

        :return: A randomly-chosen word from our list of words
        """

        return self.rand.choice(ANSWER_WORDS)

    def make_answer(self, word_count=4, sep=' '):
        """Generate and return a new answer. It's automatically added to the puzzle answer list.

        :param int word_count: The number of words to include in the answer.
        :param str|bytes sep: The word separator.
        :returns: The answer string
        """

        words = [self.randword() for i in range(word_count)]
        answer = sep.join(words)
        self.answers.append(answer)
        return answer

    hexdump_stdch = stdch = (
        '················'
        '················'
        ' !"#$%&\'()*+,-./'
        '0123456789:;<=>?'
        '@ABCDEFGHIJKLMNO'
        r'PQRSTUVWXYZ[\]^_'
        '`abcdefghijklmno'
        'pqrstuvwxyz{|}~·'
        '················'
        '················'
        '················'
        '················'
        '················'
        '················'
        '················'
        '················'
    )

    def hexdump(self, buf, charset=hexdump_stdch, gap=('�', '⌷')):
        """Write a hex dump of data to the puzzle body

        :param buf: Buffer of bytes to dump
        :param charset: Character set to use while dumping hex-equivalents. Default to ASCII
        :param gap: Length-2 tuple containing character to use to represent unprintable characters
        """
        hexes, chars = [], []
        out = []

        # Generate hex and character sequences from input buffer
        for buf_byte in buf:
            if len(chars) == 16:
                out.append((hexes, chars))
                hexes, chars = [], []

            if buf_byte is None:
                hex_char, char = gap
            else:
                hex_char = '{:02x}'.format(buf_byte)
                char = charset[buf_byte]
            chars.append(char)
            hexes.append(hex_char)

        out.append((hexes, chars))

        offset = 0
        elided = False
        lastchars = None
        self.body.write('<pre>')

        # Print out hex and character equivalents side-by-side
        for hexes, chars in out:
            if chars == lastchars:
                offset += len(chars)
                if not elided:  # pragma: no cover
                    self.body.write('*\n')
                    elided = True
                continue
            lastchars = chars[:]
            elided = False

            pad = 16 - len(chars)
            hexes += ['  '] * pad

            self.body.write('{:08x}  '.format(offset))
            self.body.write(' '.join(hexes[:8]))
            self.body.write('  ')
            self.body.write(' '.join(hexes[8:]))
            self.body.write('  |')
            self.body.write(html.escape(''.join(chars)))
            self.body.write('|\n')
            offset += len(chars)

        self.body.write('{:08x}\n'.format(offset))
        self.body.write('</pre>')

    def get_authors(self):
        """Get author names from the puzzle

        :return: List of authors
        """
        # Some legacy objects might only have self.author set
        return self.authors or [self.author] if hasattr(self, "author") else []  # pylint: disable=no-member

    def get_body(self):
        """Get the body of the puzzle

        :return: The body of the puzzle
        """
        return self.body.getvalue()

    def html_body(self):
        """Format and return the markdown for the puzzle body.

        :return: The rendered body of the puzzle
        """
        return mistune.markdown(self.get_body(), escape=False)

    def package(self, answers=False):
        """Return a dict packaging of the puzzle.

        :param answers: Whether or not to include answers in the results, defaults to False
        :return: Dict representation of the puzzle
        """

        files = [fn for fn, f in self.files.items() if f.visible]
        return {
            'answers': self.answers if answers else [],
            'authors': self.authors,
            'hashes': self.hashes(),
            'files': files,
            'scripts': self.scripts,
            'pattern': self.pattern,
            'body': self.html_body(),
        }

    def hashes(self):
        """Return a list of answer hashes

        :return: List of answer hashes
        """

        return [djb2hash(a) for a in self.answers]


class Category:
    """A category containing 1 or more puzzles

    :param path: The path to the category directory
    :param seed: The seed used for the PRNG in this category
    """
    def __init__(self, path, seed):
        path = str(path)
        self.path = path
        self.seed = seed
        self.catmod = None

        try:
            self.catmod = loadmod('category', os.path.join(path, 'category.py'))
        except FileNotFoundError:
            self.catmod = None

    def pointvals(self):
        """Return valid point values for a category

        :return: A list of valid point values for this category
        """
        if self.catmod:
            with pushd(self.path):
                pointvals = self.catmod.pointvals()
        else:
            pointvals = []
            for fpath in glob.glob(os.path.join(self.path, "[0-9]*")):
                point_name = os.path.basename(fpath)
                points = int(point_name)
                pointvals.append(points)
        return sorted(pointvals)

    def puzzle(self, points):
        """Return a puzzle with the given point value

        :param points: Point value to generate
        :return: Puzzle object worth `points` points
        """
        puzzle = Puzzle(self.seed, points)
        path = os.path.join(self.path, str(points))
        if self.catmod:
            with pushd(self.path):
                self.catmod.make(points, puzzle)
        else:
            puzzle.read_directory(path)
        return puzzle

    def __iter__(self):
        for points in self.pointvals():
            yield self.puzzle(points)
