import crypto

import itertools

width = 7

alice = b'''The key for this one was essentially 0 1 2 3 4 5 6 7.  The key for the next puzzle is much stronger.  I bet they're glad we're not also applying a substitution cypher as a secondary step.  '''
bob =   b'''I take that to mean it uses the same basic algorithm.  I guess it won't be too hard then, will it?  The key for this puzzle is this sentence'''
alice = alice.lower().replace(b' ', b'_')
bob = bob.lower().replace(b' ', b'_')

def rotate(text):
    out = bytearray()
    assert len(text) % width == 0, 'At %d of %d.' % (len(text) % width, width)

    slices = [bytearray(text[i:i+width]) for i in range(0, len(text), width)]
    nextSlice = slices.pop(0)
    while len(out) < len(text):
        if nextSlice:
            out.append(nextSlice.pop(0))
        slices.append(nextSlice)
        nextSlice = slices.pop(0)

    return bytes(out)

def unrotate(text):
    out = bytearray()
    assert len(text) % width == 0
    
    slices = []
    for i in range(len(text) // width):
        slices.append([])

    inText = bytearray(text)
    while inText:
        slice = slices.pop(0)
        slice.append(inText.pop(0))
        slices.append(slice)

    for slice in slices:
        out.extend(slice)

    return bytes(out)

print(rotate(alice))

crypto.mkIndex(rotate, unrotate, alice, bob, crypto.groups)
