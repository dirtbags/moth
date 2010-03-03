
import cbc, crypto
import diffie

alice = """Lets do a diffie hellman key exchange, Bob.  The next puzzle will be encrypted using CBC and sha512(<name>.<key>) ^ text as the cipher function, 
and an IV of 0xaa 64 times. The prime is: %d, mod: %d, and I chose %d.  Also, have some more Lovecraft: Too awed even to hint theories, the seven shaking men trudged back toward Arkham by the north road. Ammi was worse than his fellows, and begged them to see him inside his own kitchen, instead of keeping straight on to town. He did not wish to cross the nighted, wind-whipped woods alone to his home on the main road. For he had had an added shock that the others were spared, and was crushed forever with a brooding fear he dared not even mention for many years to come. As the rest of the watchers on that tempestuous hill had stolidly set their faces toward the road, Ammi had looked back an instant at the shadowed valley of desolation so lately sheltering his ill-starred friend. And from that stricken, far-away spot he had seen something feebly rise, only to sink down again upon the place from which the great shapeless horror had shot into the sky. It was just a colour—but not any colour of our earth or heavens. And because Ammi recognised that colour, and knew that this last faint remnant must still lurk down there in the well, he has never been quite right since. """ % \
(diffie.prime, diffie.mod, diffie.a)
bob = """Umm, ok.  You'll need this: %d. The key this time is 'quavering tendrils'.  Some more text to decode:  West of Arkham the hills rise wild, and there are valleys with deep woods that no axe has ever cut. There are dark narrow glens where the trees slope fantastically, and where thin brooklets trickle without ever having caught the glint of sunlight. On the gentler slopes there are farms, ancient and rocky, with squat, moss-coated cottages brooding eternally over old New England secrets in the lee of great ledges; but these are all vacant now, the wide chimneys crumbling and the shingled sides bulging perilously beneath low gambrel roofs.
The old folk have gone away, and foreigners do not like to live there. French-Canadians have tried it, Italians have tried it, and the Poles have come and departed. It is not because of anything that can be seen or heard or handled, but because of something that is imagined. The place is not good for the imagination, and does not bring restful dreams at night. It must be this which keeps the foreigners away, for old Ammi Pierce has never told them of anything he recalls from the strange days. Ammi, whose head has been a little queer for years, is the only one who still remains, or who ever talks of the strange days; and he dares to do this because his house is so near the open fields and the travelled roads around Arkham.
There was once a road over the hills and through the valleys, that ran straight where the blasted heath is now; but people ceased to use it and a new road was laid curving far toward the south. Traces of the old one can still be found amidst the weeds of a returning wilderness, and some of them will doubtless linger even when half the hollows are flooded for the new reservoir. Then the dark woods will be cut down and the blasted heath will slumber far below blue waters whose surface will mirror the sky and ripple in the sun. And the secrets of the strange days will be one with the deep’s secrets; one with the hidden lore of old ocean, and all the mystery of primal earth. 
When I went into the hills and vales to survey for the new reservoir they told me the place was evil. They told me this in Arkham, and because that is a very old town full of witch legends I thought the evil must be something which grandams had whispered to children through centuries. The name “blasted heath” seemed to me very odd and theatrical, and I wondered how it had come into the folklore of a Puritan people. Then I saw that dark westward tangle of glens and slopes for myself, and ceased to wonder at anything besides its own elder mystery. It was morning when I saw it, but shadow lurked always there. The trees grew too thickly, and their trunks were too big for any healthy New England wood. There was too much silence in the dim alleys between them, and the floor was too soft with the dank moss and mattings of infinite years of decay. """ % \
(diffie.B,)

alice = crypto.strip(alice)
bob = crypto.strip(bob)

def Ce(text, key):
    out = bytearray()
    for i in range(len(text)):
        out.append( ( (text[i] + key[i]) % 256) ^ key[i] )

    return bytes(out)

def Cd(text, key):
    out = bytearray()
    for i in range(len(text)):
        out.append( ( (text[i] ^ key[i]) - key[i]) % 256 )

    return bytes(out)

IV = b'ahiru'
key = b'space'

encode = lambda t : cbc.cipherBlockChainingE(key, IV, Ce, t)
decode = lambda t : cbc.cipherBlockChainingD(key, IV, Cd, t)

if __name__ == '__main__':
    crypto.mkIndex(encode, decode, alice, bob, crypto.hexFormat)
