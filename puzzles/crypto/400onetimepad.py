import crypto
import random

def mkPad(length):
    pad = bytearray()
    for i in range(length):
        pad.append( random.randint(0,255) )
    return bytes(pad)

alice = b'That was it, you solved the last crypto puzzle! Congratulations.  I hope you realize that, in the grand scheme of things, these were of trivial difficulty.'
bob =   b"It's not like we could expect you to solve anything actually difficult in a day, after all.       --------========Thanks for Pl@y|ng========--------       "

assert len(alice) == len(bob)
key = mkPad(len(alice))

def encode(text):
    out = bytearray()
    for i in range(len(text)):
        out.append(key[i] ^ text[i])
    return bytes(out)

crypto.mkIndex(encode, encode, alice, bob, crypto.hexFormat)
    
