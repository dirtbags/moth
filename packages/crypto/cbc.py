
def cipherBlockChainingE(key, IV, C, text):
    """Cypher block chaining encryption.  Works in blocks the size of IV.
@param key: the key for the Cipher.  
@param IV: initialization vector (bytes object).
@param C: the cypher function C(text, key). 
@param text: A bytes object of the text. The length of the text
                  must be a multiple of the length of the IV.
"""
    mod = len(text) % len(IV)
    assert mod == 0, 'The text length needs to be a multiple of the key '\
           'length.  %d of %d' % (mod, len(IV))
   
    feedback = IV
    block = len(IV)
    out = bytearray()
    while text:
        p, text = text[:block], text[block:]

        c = bytearray(block)
        for i in range(block):
            c[i] = p[i] ^ feedback[i]

        c2 = C(c, key)
        out.extend(c2)
        feedback = c2

    return bytes(out)

def cipherBlockChainingD(key, IV, C, text):
    """Cipher block chaining decryption.  Arguments are the same as for the 
encrypting function."""
    mod = len(text) % len(IV)
    assert mod == 0, 'The text length needs to be a multiple of the IV '\
           'length.  %d of %d' % (mod, len(IV))
   
    feedback = IV
    block = len(IV)
    out = bytearray()
    while text:
        c, text = text[:block], text[block:]

        p = C(c, key)
        p = bytearray(p)
        for i in range(block):
            p[i] = p[i] ^ feedback[i]

        out.extend(p)
        feedback = c

    return bytes(out)
   
