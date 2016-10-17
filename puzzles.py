#!/usr/bin/python3

import hmac
import base64
import argparse
import glob
import json
import os
import mistune
import random

messageChars = b'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'

def djb2hash(buf):
    h = 5381
    for c in buf:
        h = ((h * 33) + c) & 0xffffffff
    return h

class Puzzle:
    def __init__(self, stream):
        self.message = bytes(random.choice(messageChars) for i in range(20))
        self.fields = {}
        self.answers = []
        self.hashes = []

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
                self._add_field(key, val)
            else:
                body.append(line)
        self.body = ''.join(body)

    def _add_field(self, key, val):
        if key == 'answer':
            h = djb2hash(val.encode('utf8'))
            self.answers.append(val)
            self.hashes.append(h)
        else:
            self.fields[key] = val

    def htmlify(self):
        return mistune.markdown(self.body)

    def publish(self):
        obj = {
            'author': self.fields['author'],
            'hashes': self.hashes,
            'body': self.htmlify(),
        }
        return obj

    def secrets(self):
        obj = {
            'answers': self.answers,
            'summary': self.fields['summary'],
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
            puzzle = Puzzle(open(puzzlePath))
            puzzles[points] = puzzle

        for points in sorted(puzzles):
            puzzle = puzzles[points]
            print(puzzle.secrets())

