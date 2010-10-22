
import cbc, crypto
from transform import transform

alice = b"""You know, I just realized it's kind of smug for us to be talking about how easy or difficult these puzzles are we we're making them rather than solving them.  We've tried really hard to make them so that you don't have to follow some specific thread of logic to get to the correct answer; you just have to puzzle out the mechanism involved."""
bob = b"""The next crypto function is something simple, but new.  Here, have some more Lovecraft again: Ammi shewed them the back door and the path up through the fields to the ten-acre pasture. They walked and stumbled as in a dream, and did not dare look back till they were far away on the high ground. They were glad of the path, for they could not have gone the front way, by that well. It was bad enough passing the glowing barn and sheds, and those shining orchard trees with their gnarled, fiendish contours; but thank heaven the branches did their worst twisting high up. The moon went under some very black clouds as they crossed the rustic bridge over Chapman's Brook, and it was blind groping from there to the open meadows. open meadows """ 

IV = b'ahiru'
keyE = [2, 4, 0, 1, 3]
keyD = [2, 3, 0, 4, 1]

encode = lambda t : cbc.cipherBlockChainingE(keyE, IV, transform, t)
decode = lambda t : cbc.cipherBlockChainingD(keyD, IV, transform, t)

crypto.mkIndex(encode, decode, alice, bob, crypto.hexFormat)
