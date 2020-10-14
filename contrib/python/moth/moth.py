#!/usr/bin/python3

import argparse
import contextlib
import copy
import glob
import hashlib
import html
import io
import importlib.machinery
import logging
import os
import random
import string
import sys
import tempfile
import shlex
import yaml

from . import mistune

messageChars = b'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'

LOGGER = logging.getLogger(__name__)

def djb2hash(str):
    h = 5381
    for c in str.encode("utf-8"):
        h = ((h * 33) + c) & 0xffffffff
    return h

@contextlib.contextmanager
def pushd(newdir):
    curdir = os.getcwd()
    LOGGER.debug("Attempting to chdir from %s to %s" % (curdir, newdir))
    os.chdir(newdir)

    # Force a copy of the old path, instead of just a reference
    old_path = list(sys.path)
    old_modules = copy.copy(sys.modules)
    sys.path.append(newdir)

    try:
        yield
    finally:
        # Restore the old path
        to_remove = []
        for module in sys.modules:
            if module not in old_modules:
                to_remove.append(module)

        for module in to_remove:
            del(sys.modules[module])

        sys.path = old_path
        LOGGER.debug("Changing directory back from %s to %s" % (newdir, curdir))
        os.chdir(curdir)


def loadmod(name, path):
    abspath = os.path.abspath(path)
    loader = importlib.machinery.SourceFileLoader(name, abspath)
    return loader.load_module()


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

class PuzzleSuccess(dict):
    """Puzzle success objectives

    :param acceptable: Learning outcome from acceptable knowledge of the subject matter
    :param mastery: Learning outcome from mastery of the subject matter
    """

    valid_fields = ["acceptable", "mastery"]

    def __init__(self, **kwargs):
        super(PuzzleSuccess, self).__init__()
        for key in self.valid_fields:
            self[key] = None
        for key, value in kwargs.items():
            if key in self.valid_fields:
                self[key] = value

    def __getattr__(self, attr):
        if attr in self.valid_fields:
            return self[attr]
        raise AttributeError("'%s' object has no attribute '%s'" % (type(self).__name__, attr))

    def __setattr__(self, attr, value):
        if attr in self.valid_fields:
            self[attr] = value
        else:
            raise AttributeError("'%s' object has no attribute '%s'" % (type(self).__name__, attr))


class Puzzle:
    def __init__(self, category_seed, points):
        """A MOTH Puzzle.

        :param category_seed: A byte string to use as a seed for random numbers for this puzzle.
                              It is combined with the puzzle points.
        :param points: The point value of the puzzle.
        """

        super().__init__()

        self.points = points
        self.summary = None
        self.authors = []
        self.answers = []
        self.scripts = []
        self.pattern = None
        self.hint = None
        self.files = {}
        self.body = io.StringIO()

        # NIST NICE objective content
        self.objective = None  # Text describing the expected learning outcome from solving this puzzle, *why* are you solving this puzzle
        self.success = PuzzleSuccess()  # Text describing criteria for different levels of success, e.g. {"Acceptable": "Did OK", "Mastery": "Did even better"}
        self.solution = None  # Text describing how to solve the puzzle
        self.ksas = []  # A list of references to related NICE KSAs (e.g. K0058, . . .)

        self.logs = []
        self.randseed = category_seed * self.points
        self.rand = random.Random(self.randseed)

    def log(self, *vals):
        """Add a new log message to this puzzle."""
        msg = ' '.join(str(v) for v in vals)
        self.logs.append(msg)

    def read_stream(self, stream):
        header = True
        line = ""
        if stream.read(3) == "---":
            header = "yaml"
        else:
            header = "moth"

        stream.seek(0)

        if header == "yaml":
            LOGGER.info("Puzzle is YAML-formatted")
            self.read_yaml_header(stream)
        elif header == "moth":
            LOGGER.info("Puzzle is MOTH-formatted")
            self.read_moth_header(stream)
                
        for line in stream:
            self.body.write(line)

    def read_yaml_header(self, stream):
        contents = ""
        header = False
        for line in stream:
            if line.strip() == "---" and header:  # Handle last line
                break
            elif line.strip() == "---":  # Handle first line
                header = True
                continue
            else:
                contents += line

        config = yaml.safe_load(contents)
        for key, value in config.items():
            key = key.lower()
            self.handle_header_key(key, value)

    def read_moth_header(self, stream):
        for line in stream:
            line = line.strip()
            if not line:
                break

            key, val = line.split(':', 1)
            key = key.lower()
            val = val.strip()
            self.handle_header_key(key, val)

    def handle_header_key(self, key, val):
        LOGGER.debug("Handling key: %s, value: %s", key, val)
        if key == 'author':
            self.authors.append(val)
        elif key == 'authors':
            if not isinstance(val, list):
                raise ValueError("Authors must be a list, got %s, instead" & (type(val),))
            self.authors = list(val)
        elif key == 'summary':
            self.summary = val
        elif key == 'answer':
            if not isinstance(val, str):
                raise ValueError("Answers must be strings, got %s, instead" % (type(val),))
            self.answers.append(val)
        elif key == "answers":
            for answer in val:
                if not isinstance(answer, str):
                    raise ValueError("Answers must be strings, got %s, instead" % (type(answer),))
                self.answers.append(answer)
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
            LOGGER.debug("Attempting to open %s", os.path.abspath(name))
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
        elif key == "objective":
            self.objective = val
        elif key == "success":
            # Force success dictionary keys to be lower-case
            self.success = dict((x.lower(), y) for x,y in val.items())
        elif key == "success.acceptable":
            self.success.acceptable = val
        elif key == "success.mastery":
            self.success.mastery = val
        elif key == "solution":
            self.solution = val
        elif key == "ksas":
            if not isinstance(val, list):
                raise ValueError("KSAs must be a list, got %s, instead" & (type(val),))
            self.ksas = val
        elif key == "ksa":
            self.ksas.append(val)
        else:
            raise ValueError("Unrecognized header field: {}".format(key))


    def read_directory(self, path):
        try:
            puzzle_mod = loadmod("puzzle", os.path.join(path, "puzzle.py"))
        except FileNotFoundError:
            puzzle_mod = None

        with pushd(path):
            if puzzle_mod:
                puzzle_mod.make(self)
            else:
                with open('puzzle.moth') as f:
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

    def add_file(self, filename, visible=True):
        fd = open(filename, 'rb')
        name = os.path.basename(filename)
        self.add_stream(fd, name=name, visible=visible)

    def randword(self):
        """Return a randomly-chosen word"""

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
        'PQRSTUVWXYZ[\]^_'
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
        hexes, chars = [], []
        out = []

        for b in buf:
            if len(chars) == 16:
                out.append((hexes, chars))
                hexes, chars = [], []

            if b is None:
                h, c = gap
            else:
                h = '{:02x}'.format(b)
                c = charset[b]
            chars.append(c)
            hexes.append(h)

        out.append((hexes, chars))

        offset = 0
        elided = False
        lastchars = None
        self.body.write('<pre>')
        for hexes, chars in out:
            if chars == lastchars:
                offset += len(chars)
                if not elided:
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
        return self.authors or [self.author]

    def get_body(self):
        return self.body.getvalue()

    def html_body(self):
        """Format and return the markdown for the puzzle body."""
        return mistune.markdown(self.get_body(), escape=False)

    def package(self, answers=False):
        """Return a dict packaging of the puzzle."""

        files = [fn for fn,f in self.files.items() if f.visible]
        return {
            'authors': self.get_authors(),
            'hashes': self.hashes(),
            'files': files,
            'scripts': self.scripts,
            'pattern': self.pattern,
            'body': self.html_body(),
            'objective': self.objective,
            'success': self.success,
            'solution': self.solution,
            'ksas': self.ksas,
        }

    def hashes(self):
        "Return a list of answer hashes"

        return [djb2hash(a) for a in self.answers]


class Category:
    def __init__(self, path, seed):
        self.path = path
        self.seed = seed
        self.catmod = None

        try:
            self.catmod = loadmod('category', str(os.path.join(str(path), 'category.py')))
        except FileNotFoundError:
            self.catmod = None

    def pointvals(self):
        if self.catmod:
            with pushd(self.path):
                pointvals = self.catmod.pointvals()
        else:
            pointvals = []
            for fpath in glob.glob(str(os.path.join(str(self.path), "[0-9]*"))):
                pn = os.path.basename(fpath)
                points = int(pn)
                pointvals.append(points)
        return sorted(pointvals)

    def puzzle(self, points):
        puzzle = Puzzle(self.seed, points)
        path = str(os.path.join(str(self.path), str(points)))
        if self.catmod:
            with pushd(self.path):
                self.catmod.make(points, puzzle)
        else:
            with pushd(str(self.path)):
                puzzle.read_directory(path)
        return puzzle

    def __iter__(self):
        for points in self.pointvals():
            yield self.puzzle(points)
