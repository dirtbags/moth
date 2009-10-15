import crypto

alpha = b'abcdefghiklmnoprstuw'

alice = b'''The next four puzzles are all transposition cyphers like this one.  Transposition, like substition, is still used in modern crypto systems.     '''
bob = b'''Transposition cyphers often work with the text arranged into blocks of a certain width, often as determined by the key.  Dangling parts are often padded with nulls or random text. terrifying silence  '''
alice = alice.replace(b' ', b'_').lower()
bob = bob.replace(b' ', b'_').lower()

map = [6, 3, 0, 5, 2, 7, 4, 1]
imap = [2, 7, 4, 1, 6, 3, 0, 5] 


encode = lambda t : transform(t, map)
decode = lambda t : transform(t, imap)

crypto.mkIndex(encode, decode, alice, bob, crypto.groups)


