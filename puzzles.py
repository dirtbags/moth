#!/usr/bin/python3

import argparse
from collections import defaultdict, namedtuple
import glob
import hashlib
from importlib.machinery import SourceFileLoader
import mistune
import os
import random
import tempfile

messageChars = b'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'


def djb2hash(buf):
    h = 5381
    for c in buf:
        h = ((h * 33) + c) & 0xffffffff
    return h

# We use a named tuple rather than a full class, because any random name generation has
# to be done with Puzzle's random number generator, and it's cleaner to not pass that around.
PuzzleFile = namedtuple('PuzzleFile', ['path', 'handle', 'name', 'visible'])

class Puzzle:

    KNOWN_KEYS = [
        'file',
        'resource',
        'temp_file',
        'answer',
        'points',
        'author',
        'summary'
    ]
    REQUIRED_KEYS = [
        'author',
        'answer',
        'points'
    ]
    SINGULAR_KEYS = [
        'points'
    ]

    # Get a big list of clean words for our answer file.
    ANSWER_WORDS = [w.strip() for w in open(os.path.join(os.path.dirname(__file__),
                                                         'answer_words.txt'))]

    def __init__(self, path, category_seed):
        super().__init__()

        self._dict = defaultdict(lambda: [])
        if os.path.isdir(path):
            self._puzzle_dir = path
        else:
            self._puzzle_dir = None
        self.message = bytes(random.choice(messageChars) for i in range(20))
        self.body = ''

        if not os.path.exists(path):
            raise ValueError("No puzzle at path: {]".format(path))
        elif os.path.isfile(path):
            try:
                # Expected format is path/<points_int>.moth
                self['points'] = int(os.path.split(path)[-1].split('.')[0])
            except (IndexError, ValueError):
                raise ValueError("Invalid puzzle config. "
                                 "Expected something like <point_value>.moth")

            stream = open(path)
            self._read_config(stream)
        elif os.path.isdir(path):
            try:
                # Expected format is path/<points_int>.moth
                self['points'] = int(os.path.split(path)[-1])
            except (IndexError, ValueError):
                raise ValueError("Invalid puzzle config. Expected an integer point value for a "
                                 "directory name.")

            files = os.listdir(path)

            if 'config.moth' in files:
                self._read_config(open(os.path.join(path, 'config.moth')))

            if 'make.py' in files:
                # Good Lord this is dangerous as fuck.
                loader = SourceFileLoader('puzzle_mod', os.path.join(path, 'make.py'))
                puzzle_mod = loader.load_module()
                if hasattr(puzzle_mod, 'make'):
                    puzzle_mod.make(self)
        else:
            raise ValueError("Unacceptable file type for puzzle at {}".format(path))

        self._seed = hashlib.sha1(category_seed + bytes(self['points'])).digest()
        self.rand = random.Random(self._seed)

        # Set our 'files' as a dict, since we want register them uniquely by name.
        self['files'] = dict()

    def _read_config(self, stream):
        """Read a configuration file (ISO 2822)"""
        body = []
        header = True
        for line in stream:
            if header:
                line = line.strip()
                if not line.strip():
                    header = False
                    continue
                key, val = line.split(':', 1)
                key = key.lower()
                val = val.strip()
                self[key] = val
            else:
                body.append(line)
        self.body = ''.join(body)

    def random_hash(self):
        """Create a random hash from our number generator suitable for use as a filename."""
        return hashlib.sha1(str(self.rand.random()).encode('ascii')).digest()

    def _puzzle_file(self, path, name, visible=True):
        """Make a puzzle file instance for the given file.
        :param path: The path to the file
        :param name: The name of the file. If set to None, the published file will have
                     a random hash as a name and have visible set to False.
        :return:
        """

        # Make sure it actually exists.
        if not os.path.exists(path):
            raise ValueError("Included file {} does not exist.")

        file = open(path, 'rb')

        return PuzzleFile(path=path, handle=file, name=name, visible=visible)

    def make_file(self, name=None, mode='rw+b'):
        """Get a file object for adding dynamically generated data to the puzzle.
        :param name: The name of the file for links within the puzzle. If this is None,
        the file will be hidden with a random hash as the name.
        :return: A file object for writing
        """

        file = tempfile.TemporaryFile(mode=mode, delete=False)

        self._dict['files'].append(self._puzzle_file(file.name, name))

        return file

    def __setitem__(self, key, value):

        if key in ('file', 'resource', 'hidden') and self._puzzle_dir is None:
            raise KeyError("Cannot set a puzzle file for single file puzzles.")

        if key == 'answer':
            # Handle adding answers to the puzzle
            self._dict['hashes'].append(djb2hash(value.encode('utf8')))
            self._dict['answers'].append(value)
        elif key == 'file':
            # Handle adding files to the puzzle
            path = os.path.join(self._puzzle_dir, 'files', value)
            self._dict['files'][value] = self._puzzle_file(path, value)
        elif key == 'resource':
            # Handle adding category files to the puzzle
            path = os.path.join(self._puzzle_dir, '../res', value)
            self._dict['files'].append(self._puzzle_file(path, value))
        elif key == 'hidden':
            # Handle adding secret, 'hidden' files to the puzzle.
            path = os.path.join(self._puzzle_dir, 'files', value)
            name = self.random_hash()
            self._dict['files'].append(self._puzzle_file(path, name, visible=False))
        elif key in self.SINGULAR_KEYS:
            # These keys can only have one value
            self._dict[key] = value
        elif key in self.KNOWN_KEYS:
            self._dict[key].append(value)
        else:
            raise KeyError("Invalid Attribute: {}".format(key))

    def __getitem__(self, item):
        return self._dict[item]

    def make_answer(self, word_count, sep=b' '):
        """Generate and return a new answer. It's automatically added to the puzzle answer list.
        :param int word_count: The number of words to include in the answer.
        :param str|bytes sep: The word separator.
        :returns: The answer bytes
        """

        if type(sep) == str:
            sep = sep.encode('ascii')

        answer = sep.join(self.rand.sample(self.ANSWER_WORDS, word_count))
        self['answer'] = answer

        return answer


    def htmlify(self):
        return mistune.markdown(self.body)

    def publish(self, dest):
        """Deploy the puzzle to the given directory, and return the info needed for describing
        the puzzle and accepting answers in MOTH."""

        if not os.path.exists(dest):
            raise ValueError("Puzzle destination directory does not exist.")

        # Delete the original directory

        # Save puzzle html file

        # Copy over all the files.

        obj = {
            'author': self['author'],
            'hashes': self['hashes'],
            'body': self.htmlify(),
        }
        return obj

    def secrets(self):
        obj = {
            'answers': self['answers'],
            'summary': self['summary'],
        }
        return obj
    
if __name__ == '__main__':        
    parser = argparse.ArgumentParser(description='Build a puzzle category')
    parser.add_argument('puzzledir', nargs='+', help='Directory of puzzle source')
    args = parser.parse_args()

    for puzzledir in args.puzzledir:
        puzzles = {}
        secrets = {}
        for puzzlePath in glob.glob(os.path.join(puzzledir, "*.moth")):
            filename = os.path.basename(puzzlePath)
            points, ext = os.path.splitext(filename)
            points = int(points)
            puzzle = Puzzle(puzzlePath, "test")
            puzzles[points] = puzzle

        for points in sorted(puzzles):
            puzzle = puzzles[points]
            print(puzzle.secrets())

