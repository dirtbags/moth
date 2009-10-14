import crypto

alpha = b'abcdefghiklmnoprstuw'

alice = b'''The next four puzzles are all transposition cyphers like this one.  Transposition, like substition, is still used in modern crypto systems.     '''
bob = b'''Transposition cyphers often work with the text arranged into blocks of a certain width, often as determined by the key.  Dangling parts are often padded with nulls or random text. terrifying silence  '''
alice = alice.replace(b' ', b'_').lower()
bob = bob.replace(b' ', b'_').lower()

map = [6, 3, 0, 5, 2, 7, 4, 1]
imap = [2, 7, 4, 1, 6, 3, 0, 5] 

def transform(text, map):
    assert len(text) % 8 == 0, 'Text must be multiple of 8 in length.  '\
                               '%d more chars needed.' % (8 - len(text) % 8)

    out = bytearray()
    i = 0
    while i < len(text):
        for j in range(8):
            out.append( text[i + map[j]] )
        i = i+8
    return bytes(out)

encode = lambda t : transform(t, map)
decode = lambda t : transform(t, imap)

crypto.mkIndex(encode, decode, alice, bob, crypto.groups)


