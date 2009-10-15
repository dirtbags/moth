def mkIndex(encode, decode, alice, bob, 
            format=lambda s: str(s, 'utf-8')):
    """Write out the index.html contents.
@param encode: function to encrypt the plaintext
@param decode: function to decrypt the plaintext
@param alice: plaintext of alice line
@param bob: plaintext of bob line
@param format: formatter for the cypher text, run out output of encode before
               printing. Does string conversion by default."""
    c = encode(alice)
    print('<dl><dt>Alice<dd>', format(c))
    assert decode(c) == alice
    c = encode(bob)
    print('<dt>Bob<dd>', format(c), '</dl>')
    assert decode(c) == bob

def hexFormat(text):
    return groups(text, 5, '{0:x} ')

def groups(text, perLine=5, format='{0:c}'):
    i = 0
    out = []
    while i < len(text):
        out.append(format.format(text[i]))
        
        if i % (perLine*5) == (perLine * 5 - 1):
            out.append('<BR>')
        elif i % 5 == 4:
            out.append(' ')

        i = i + 1

    return ''.join(out)
   
def strip(text):
    """Strip any unicode from the given text, and return it as a bytes 
    object."""

    b = bytearray()
    for t in text:
        if ord(t) > 255:
            t = ' '

        b.append(ord(t))

    return bytes(b)
