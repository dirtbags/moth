#!/usr/bin/python3
"""This is morris code, except the dots and dashes are each represented by
many different possible characters.  The 'encryption key' is the set of
characters that represent dots, and the set that represents dashes."""

import random

dots =   b'acdfhkjnpsrtx'
dashes = b'begilmoquvwyz'

morris = {'a': '.-',
          'b': '-...',
          'c': '-.-.',
          'd': '-..',
          'e': '.',
          'f': '..-.',
          'g': '--.',
          'h': '....',
          'i': '..',
          'j': '.---',
          'k': '-.-',
          'l': '.-..',
          'm': '--',
          'n': '-.',
          'o': '---',
          'p': '.--.',
          'q': '--.-',
          'r': '.-.',
          's': '...',
          't': '-',
          'u': '..-',
          'v': '...-',
          'w': '.--',
          'x': '-..-',
          'y': '-.--',
          'z': '--..',
          '.': '._._._',
          ',': '--..--',
          ':': '---...'}

imorris = {}
for k in morris:
    imorris[morris[k]] = k            

plaintext = [b'It is fun to make up bizarre cyphers, but the next one is '
             b'something a little more standard.',
             b'All I have to say is: giant chickens.']


def encode(text):
    out = bytearray()
    for t in text:
        if t == ord(' '):
            out.append('  ')
        else:
            for bit in morris[chr(t)]:
                if bit == '.':
                    out.append(random.choice(dots))
                else:
                    out.append(random.choice(dashes))
            out.append(' ')
    return bytes(out)

def decode(text):
    text = text.replace(b'  ', b'&')
    words = text.split(b'&')
    out = bytearray()
    for word in words:
        for c in word.split(' '):
            

c = encode(plaintext[0])
print('<dl><dt>Alice<dd>', str(c, 'utf-8'))
assert decode(c) == plaintext[0]
c = encode(plaintext[1])
print('<dt>Bob<dd>', str(c, 'utf-8'), '</dl>')
assert decode(c) == plaintext[1]
