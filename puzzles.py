#!/usr/bin/python3

import argparse
import base64
import glob
import hmac
import json
import mistune
import multidict
import os
import random

messageChars = b'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'

def djb2hash(buf):
    h = 5381
    for c in buf:
        h = ((h * 33) + c) & 0xffffffff
    return h

class Puzzle(multidict.MultiDict):

    def __init__(self, seed):
        super().__init__()

        self.message = bytes(random.choice(messageChars) for i in range(20))
        self.body = ''

        self.rand = random.Random(seed)

    @classmethod
    def from_stream(cls, stream):
        pzl = cls(None)

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
                pzl.add(key, val)
            else:
                body.append(line)
        pzl.body = ''.join(body)
        return pzl

    def add(self, key, value):
        super().add(key, value)
        if key == 'answer':
            super().add(hash, djb2hash(value.encode('utf8')))

    def htmlify(self):
        return mistune.markdown(self.body)

    def publish(self):
        obj = {
            'author': self['author'],
            'hashes': self.getall('hash'),
            'body': self.htmlify(),
        }
        return obj

    def secrets(self):
        obj = {
            'answers': self.getall('answer'),
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
            puzzle = Puzzle.from_stream(open(puzzlePath))
            puzzles[points] = puzzle

        for points in sorted(puzzles):
            puzzle = puzzles[points]
            print(puzzle.secrets())

