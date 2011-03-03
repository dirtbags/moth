#! /usr/bin/python3

import random
import array
import sys

SIZE = 2048

substrate = array.array('B', (random.randrange(256) for i in range(SIZE)))

key = open('key', 'rb').read().strip()

index = array.array('H')
for i in key:
    while True:
        pos = random.randrange(SIZE)
        if pos not in index:
            break
    index.append(pos)
index.append(0)

outbytes = index.tostring() + substrate
out = array.array('B', outbytes[:SIZE])

for i in range(len(key)):
    out[index[i]] = key[i]

sys.stdout.buffer.write(out)
