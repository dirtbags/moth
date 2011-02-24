#! /usr/bin/python3

import sys
import struct
import random

class Container:
    def __init__(self):
        self.contents = []

    def add(self, opcode, subcode, part, text):
        hdr = struct.pack('!BBHH',
                          opcode, subcode, part,
                          len(text))
        self.contents.append(hdr + text)
        random.shuffle(self.contents)

    def bytes(self):
        body = b''.join(self.contents)
        hdr = struct.pack('!LHL',
                          0xB00FB00F,
                          2,
                          len(body))
        return hdr + body



c = Container()
s = open(sys.argv[1], 'rb')
i = 0
while True:
    b = s.read(3150)
    if not b:
        break
    c.add(5, 8, i, b)
    i += 1
sys.stdout.buffer.write(c.bytes())
