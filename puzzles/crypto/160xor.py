#!/usr/bin/python3

import crypto

alice = b'''Do you think they'll try another frequency count?  It might be better if they just looked for patterns.'''
bob = b'''You'd be amazed at how often this is used in lieu of real crypto.  It's about as effective as a ceasar cypher.  chronic failure'''

key = 0xac

def encode(text):
    out = bytearray()
    for t in text:
        out.append(t ^ key)
    return bytes(out)

crypto.mkIndex(encode, encode, alice, bob, crypto.hexFormat)
