#!/usr/bin/python3

plaintext = [b"This may seem relatively simple, but it's the same basic "
b"principles as used in s-boxes, a technique used in many modern "
b"cryptographic algoritms.  Of course, instead of letter substitution, "
b"you're doing byte substitution.",
b"The next two puzzles are a bit different;  Frequency counts (of characters) "
b"will just reveal random noise. Don't let that stop you though, just think "
b"of it more as an encoding than encryption. "
b"Oh, by the way, the key this time is: 'the s is for sucks'."]

key = b"thequickbrownfxjmpdvlazygs"

def encode(text):
    ukey = key.upper()
    lkey = key.lower()
    assert len(set(key)) == 26, 'invalid key'
    assert key.isalpha(), 'non alpha character in key'
    out = bytearray()
    for t in text:
        if t in lkey:
            out.append(lkey[t - ord('a')])
        elif t in ukey:
            out.append(ukey[t - ord('A')])
        else:
            out.append(t)
    return bytes(out)
    
def decode(text):
    ukey = key.upper()
    lkey = key.lower()
    assert len(set(key)) == 26, 'invalid key'
    assert key.isalpha(), 'non alpha character in key'
    out = bytearray()
    for t in text:
        if t in lkey:
            out.append(ord('a') + lkey.index(bytes([t])))
        elif t in ukey: 
            out.append(ord('A') + ukey.index(bytes([t])))
        else:
            out.append(t)
    return bytes(out)

c = encode(plaintext[0])
print('<dl><dt>Alice<dd>', str(c, 'utf-8'))
assert decode(c) == plaintext[0]
c = encode(plaintext[1])
print('<dt>Bob<dd>', str(c, 'utf-8'), '</dl>')
assert decode(c) == plaintext[1]
