#!/usr/bin/python3
"""This is non-obvious, so let me elaborate.  The message is translated to 
binary with one character per binary bit.  Lower case characters are 1's, 
and upper case is 0.  The letters are chosen at random.  Tricky, eh?"""

import random
import crypto

lower = b'abcdefghijklmnopqrstuvwxyz'
upper = lower.upper()

plaintext = [b'The next puzzle starts in the same way, but the last step is '
             b'different.',
             b'We wouldn\'t want them to get stuck so early, would we? '
             b'Rat Fink']

def encode(text):
    out = bytearray()
    mask = 0x80
    for t in text:
        for i in range(8):
            if t & mask:
                out.append(random.choice(lower))
            else:
                out.append(random.choice(upper))
            t = t << 1
   
    return bytes(out)

def decode(text):
    out = bytearray()
    i = 0
    while i < len(text):
        c = 0
        mask = 0x80
        for j in range(8):
            if text[i] in lower:
                c =  c + mask
            mask = mask >> 1
            i = i + 1
        out.append(c)
    return bytes(out)

crypto.mkIndex(encode, decode, plaintext[0], plaintext[1], crypto.groups)
