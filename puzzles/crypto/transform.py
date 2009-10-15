def transform(text, map):
    size = len(map)
    div = len(text) % size
    assert div == 0, 'Text must be a multiple of the key size in length. '\
                     'At %d out of %d' % (div, size)

    out = bytearray()
    i = 0
    while i < len(text):
        for j in range(size):
            out.append( text[i + map[j]] )
        i = i+size
    return bytes(out)
