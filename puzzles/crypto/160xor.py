#!/usr/bin/python3

plaintext = [b'I wonderr if they'll try doing a frequency count again? '
             b'It should work this time as well.  Hopefully messing around '
             b'with simple cyphers like


for p in plaintext:
    c = sbox(text, key)
    assert c == sbox(c, ikey), 'Failure'
    print(c)
