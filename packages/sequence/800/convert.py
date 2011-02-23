#! /usr/bin/python

import struct
import random
import sys

def packet(seq, txt):
    return struct.pack('!HB', seq, len(txt)) + txt

key = open("key").read().strip()

i = 0
seq = 83
packets = []
while i < len(key):
    l = random.randrange(5)+1
    p = packet(seq, key[i:i+l])
    packets.append(p)
    i += l
    seq += 1

random.shuffle(packets)
sys.stdout.write(''.join(packets))

