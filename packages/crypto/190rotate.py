import crypto

import itertools

width = 5

alice = b'''If we did the morris code encoding prior to this transposition, I don't think anyone would ever figure out the solution.'''
bob =   b'''That's basically true of the combination of many of these techniques.  Combining a substitution along with a permutation (or transposition) satisfies the Shannon's diffusion principle of cryptography; you want to try to get rid of as much statistical information as possible.  statistical information'''
alice = alice.lower().replace(b' ', b'_')
bob = bob.lower().replace(b' ', b'_')

key = [4,2,3,1,0]

def rotate(text):
    out = bytearray()
    assert len(text) % width == 0, 'At %d of %d.' % (len(text) % width, width)

    slices = [bytearray(text[i:i+width]) for i in range(0, len(text), width)]
    for i in range(width):
        for slice in slices:
            out.append(slice[key[i]])

    return bytes(out)

def unrotate(text):
    out = bytearray()
    assert len(text) % width == 0
    
    # Make column slices, and rearrange them according to the key.
    size = len(text) // width
    tSlices = [bytearray(text[i*size:i*size+size]) for i in range(width)]
    slices = [None] * width
    for i in range(width):
        slices[key[i]] = tSlices[i]

    while len(out) < len(text):
        for i in range(5):
            out.append(slices[i].pop(0))

    return bytes(out)

crypto.mkIndex(rotate, unrotate, alice, bob, crypto.groups)
