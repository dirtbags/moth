#! /usr/bin/python3

import sys
import struct
import random
stdch = ('␀·········␊··␍··'
         '················'
         ' !"#$%&\'()*+,-./'
         '0123456789:;<=>?'
         '@ABCDEFGHIJKLMNO'
         'PQRSTUVWXYZ[\]^_'
         '`abcdefghijklmno'
         'pqrstuvwxyz{|}~·'
         '················'
         '················'
         '················'
         '················'
         '················'
         '················'
         '················'
         '················')

def hexdump(buf, fd=sys.stdout, charset=stdch):
    offset = 0
    last = None
    elided = False
    for offset in range(0, len(buf), 16):
        l = buf[offset:offset+16]

        if l == last:
            if not elided:
                fd.write("*\n")
                elided = True
            continue
        else:
            last = l
            elided = False

        pad = 16-len(l)

        hx = []
        for b in l:
            hx.append('%02x' % b)
        hx += ['  '] * pad

        fd.write('%08x  ' % offset)
        fd.write(' '.join(hx[:8]))
        fd.write('  ')
        fd.write(' '.join(hx[8:]))
        fd.write('  |')
        fd.write(''.join(charset[b] for b in l))
        fd.write('|\n')
    fd.write('%08x\n' % len(buf))


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
s = open('salad.jpg', 'rb')
i = 0
while True:
    b = s.read(3150)
    if not b:
        break
    c.add(5, 8, i, b)
    i += 1
hexdump(c.bytes())
