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
          '.': '.-.-.-',
          ',': '--..--',
          ':': '---...'}

imorris = {}
for k in morris:
    imorris[morris[k]] = k            

plaintext = [b'it is fun to make up bizarre cyphers, but the next one is '
             b'something a little more standard.',
             b'all i have to say is: giant chickens.']



def encode(text):
    out = bytearray()
    for t in text:
        if t == ord(' '):
            out.extend(b'  ')
        else:
            for bit in morris[chr(t)]:
                if bit == '.':
                    out.append(random.choice(dots))
                else:
                    out.append(random.choice(dashes))
            out.append(ord(' '))
    return bytes(out[:-1])

def decode(text):
    text = text.replace(b'   ', b'&')
#    print(text)
    words = text.split(b'&')
    out = bytearray()
    for word in words:
#        print(word)
        word = word.strip()
        for parts in word.split(b' '):
            code = []
            for part in parts:
                if part in dots:
                    code.append('.')
                else:
                    code.append('-')
            code = ''.join(code)
            out.append(ord(imorris[code]))
        out.append(ord(' '))
    return bytes(out[:-1])

c = encode(plaintext[0])
print('<dl><dt>Alice<dd>', str(c, 'utf-8'))
assert decode(c) == plaintext[0]
c = encode(plaintext[1])
print('<dt>Bob<dd>', str(c, 'utf-8'), '</dl>')
assert decode(c) == plaintext[1]
