plaintext = [b'all hail caesar.', b'caesar is the key']

alpha = b'abcdefghijklmnopqrstuvwxyz'

def ceasar(text, r):
    out = bytearray()
    for t in text:
        if t in alpha:
            t = t - b'a'[0]
            t = (t + r)%26
            out.append(t + b'a'[0])
        else:
            out.append(t)
    return bytes(out)

encode = lambda text : ceasar(text, 13)
decode = lambda text : ceasar(text, -13)

c = encode(plaintext[0])
print('<dl><dt>Alice<dd>', str(c, 'utf-8'))
assert decode(c) == plaintext[0]
c = encode(plaintext[1])
print('<dt>Bob<dd>', str(c, 'utf-8'), '</dl>')
assert decode(c) == plaintext[1]
