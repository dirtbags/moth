
import cbc, crypto

alice = b"""Do you think they've figured out that this was encrypted using cipher block chaining with a cipher of C(key, text) = text?  If they somehow stumbled across the solution with knowing what it was, the next three will be hard.  """
bob = b"""Well, either way, we might as well let them know that the next three puzzles all uses CBC, but with progressively more difficult cipher functions.  the squirrels crow at noon """

def C(text, key):
    return text

IV = b'ahiru'
key = None

encode = lambda t : cbc.cipherBlockChainingE(key, IV, C, t)
decode = lambda t : cbc.cipherBlockChainingD(key, IV, C, t)

crypto.mkIndex(encode, decode, alice, bob, crypto.hexFormat)
