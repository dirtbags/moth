import crypto
import cbc
import diffie
import hashlib

IV = [0xaa]*64
aliceKey = hashlib.sha512(bytes('alice.%d' % diffie.key, 'utf-8')).digest()
bobKey = hashlib.sha512(bytes('bob.%d' % diffie.key, 'utf-8')).digest()

alice = b"""Only one more puzzle to go.  They'll never get it though, since we use a one time pad. I need to add more text here to pad this."""
bob = b"""I wouldn't be so sure of that.    The key is:  in the same vein """

def C(text, key):
    out = bytearray()
    for i in range( len( text ) ):
        out.append(key[i] ^ text[i])

    return bytes(out)

c = cbc.cipherBlockChainingE(aliceKey, IV, C, alice)
print('<dl><dt>Alice<dd>', crypto.hexFormat(c))
assert cbc.cipherBlockChainingD(aliceKey, IV, C, c) == alice
c = cbc.cipherBlockChainingE(bobKey, IV, C, bob)
assert cbc.cipherBlockChainingD(bobKey, IV, C, c) == bob
print('<dt>Bob<dd>', crypto.hexFormat(c), '</dl>')
