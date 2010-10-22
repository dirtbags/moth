
import cbc, crypto

alice = b"""I'd say this was easy, but we didn't give them the key, did we? I'm adding some text to give them something to work with: Commencing his descent of the dark stairs, Ammi heard a thud below him. He even thought a scream had been suddenly choked off, and recalled nervously the clammy vapour which had brushed by him in that frightful room above. What presence had his cry and entry started up? Halted by some vague fear, he heard still further sounds below. Indubitably there was a sort of heavy dragging, and a most detestably sticky noise as of some fiendish and unclean species of suction. With an associative sense goaded to feverish heights, he thought unaccountably of what he had seen upstairs. Good God! What eldritch dream-world was this into which he had blundered? He dared move neither backward nor forward, but stood there trembling at the black curve of the boxed-in staircase. Every trifle of the scene burned itself into his brain. The sounds, the sense of dread expectancy, the darkness, the steepness of the narrow steps-and merciful heaven! . . . the faint but unmistakable luminosity of all the woodwork in sight; steps, sides, exposed laths, and beams alike!  """
bob = b"""No, and they\'ll have to figure out the key for the next one too.  Here\'s some Lovecraft: "Nothin\' . . . nothin\' . . . the colour . . . it burns . . . cold an\' wet . . . but it burns . . . it lived in the well . . . I seen it . . . a kind o\' smoke . . . jest like the flowers last spring . . . the well shone at night . . . Thad an\' Mernie an\' Zenas . . . everything alive . . . suckin\' the life out of everything . . . in that stone . . . it must a\' come in that stone . . . pizened the whole place . . . dun\'t know what it wants . . . that round thing them men from the college dug outen the stone . . . they smashed it . . . it was that same colour . . . jest the same, like the flowers an\' plants . . . must a\' ben more of \'em . . . seeds . . . seeds . . . they growed . . . I seen it the fust time this week . . . must a\' got strong on Zenas . . . he was a big boy, full o\' life . . . it beats down your mind an\' then gits ye . . . burns ye up . . . in the well water . . . you was right about that . . . evil water . . . Zenas never come back from the well . . . can\'t git away . . . draws ye . . . ye know summ\'at\'s comin\', but \'tain\'t no use . . . I seen it time an\' agin senct Zenas was took . . . whar\'s Nabby, Ammi? . . . my head\'s no good . . . dun\'t know how long senct I fed her . . . it\'ll git her ef we ain\'t keerful . . . jest a colour . . . her face is gettin\' to hev that colour sometimes towards night . . . an\' it burns an\' sucks . . . it come from some place whar things ain\'t as they is here . . . one o\' them professors said so . . . he was right . . . look out, Ammi, it\'ll do suthin\' more . . . sucks the life out. . . ." The Colour Out of Space"""

def C(text, key):
    out = bytearray()
    for i in range(len(text)):
        out.append( text[i] ^ key[i] )

    return bytes(out)

IV = b'ahiru'
key = b'color'

encode = lambda t : cbc.cipherBlockChainingE(key, IV, C, t)
decode = lambda t : cbc.cipherBlockChainingD(key, IV, C, t)

crypto.mkIndex(encode, decode, alice, bob, crypto.hexFormat)
