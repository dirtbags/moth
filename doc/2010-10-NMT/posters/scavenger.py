#! /usr/bin/python3

import random
import binascii

def mask(buf1, buf2):
    return bytes(a^b for (a,b) in zip(buf1, buf2))

ptext = b"xecip-nvkop-zogyr-manef-voxyx"

pads = [b"Good job figuring out the hex",
        b"encoding, but there's more to",
        b"it!  Bring your result to the",
        b"contest for points!          "]

t = ptext
for p in pads:
    print(binascii.hexlify(p).decode())
    t = mask(p, t)

print(binascii.hexlify(t).decode())
