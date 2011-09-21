#! /usr/bin/python3

import binascii
import sys

def mask(buf1, buf2):
    return bytes(a^b for (a,b) in zip(buf1, buf2))

t = bytes([0]*29)
for line in sys.stdin:
    line = line.strip().encode()
    a = binascii.unhexlify(line)
    t = mask(t, a)
print(t)
