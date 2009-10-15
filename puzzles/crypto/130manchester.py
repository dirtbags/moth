#!/usr/bin/python3
'''This is the same as the previous, but it uses non-return to zero to encode
the binary.'''

import random

lower = b'abcdefghijklmnopqrstuvwxyz'
upper = lower.upper()

plaintext = [b'The next one is in Morris Code.  Unlike the previous two, '
             b'they will actually need to determine some sort of key.',
             b'Morris code with a key?  That sounds bizarre. probable cause']

def encode(text):
    out = bytearray()
    mask = 0x80
    state = 0
    for t in text:
        for i in range(8):
            next = t & mask
            if not state and not next:
                out.append(random.choice(upper))
                out.append(random.choice(lower))
            elif not state and next:
                out.append(random.choice(lower))
                out.append(random.choice(upper))
            elif state and not next:
                out.append(random.choice(upper))
                out.append(random.choice(lower))
            elif state and next:
                out.append(random.choice(lower))
                out.append(random.choice(upper))
            state = next
            t = t << 1
   
    return bytes(out)

def decode(text):
    out = bytearray()
    i = 0
    while i < len(text):
        c = 0
        mask = 0x80
        for j in range(8):
            a = 0 if text[i] in lower else 1
            b = 0 if text[i+1] in lower else 1
            assert a != b, 'bad encoding'
            if b:
                c =  c + mask
            mask = mask >> 1
            i = i + 2
        out.append(c)
    return bytes(out)

c = encode(plaintext[0])
print('<dl><dt>Alice<dd>', str(c, 'utf-8'))
assert decode(c) == plaintext[0]
c = encode(plaintext[1])
print('<dt>Bob<dd>', str(c, 'utf-8'), '</dl>')
assert decode(c) == plaintext[1]
