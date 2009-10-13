#!/usr/bin/python3

key = [43, 44, 227, 31, 255, 42, 194, 197, 187, 11, 92, 234, 57, 67, 45, 40, 66, 226, 214, 184, 167, 139, 210, 233, 22, 246, 150, 75, 186, 145, 86, 224, 17, 131, 24, 98, 74, 248, 213, 212, 72, 101, 160, 221, 243, 69, 113, 142, 127, 47, 141, 68, 247, 138, 124, 177, 192, 165, 110, 107, 203, 207, 254, 176, 154, 8, 87, 189, 228, 155, 143, 0, 220, 1, 128, 3, 169, 204, 162, 90, 156, 208, 170, 222, 95, 223, 188, 215, 174, 78, 48, 50, 244, 116, 179, 134, 171, 153, 15, 196, 135, 52, 85, 195, 71, 32, 190, 191, 21, 161, 63, 218, 64, 106, 123, 239, 235, 241, 34, 61, 144, 152, 111, 20, 172, 117, 237, 120, 80, 88, 200, 185, 109, 137, 37, 159, 183, 30, 202, 129, 250, 58, 9, 193, 41, 164, 65, 126, 46, 158, 132, 97, 166, 6, 23, 147, 105, 29, 38, 119, 76, 238, 240, 12, 201, 245, 230, 14, 206, 114, 10, 25, 60, 83, 236, 18, 231, 39, 77, 55, 252, 229, 100, 7, 28, 209, 51, 148, 181, 198, 225, 118, 173, 103, 35, 149, 91, 108, 219, 168, 140, 49, 33, 122, 82, 216, 53, 205, 13, 73, 249, 180, 81, 19, 112, 232, 217, 96, 62, 99, 4, 26, 178, 211, 199, 151, 102, 121, 253, 136, 130, 104, 133, 146, 89, 5, 157, 70, 84, 242, 182, 93, 251, 54, 16, 175, 56, 115, 94, 36, 27, 79, 59, 163, 125, 2]
ikey = [None]*256
for i in range(256):
    ikey[key[i]] = i

plaintext = [b'I think it's impressive if they get this one.  It will take a '
	     b'lot of work to get it right.  That is, unless they do '
             b'something smart like correctly guess the value of spaces. '
             b'Frequency counts won't just be your friend here, it'll be '
             b'useful in other places too.',
	     b'I'm not sure if that's enough text to give them the '
	     b'ability to make a good frequency count.  It's nice to '
	     b'finally be at a real cypher that allows for things like '
             b'proper punctuation.  Anyway, the key is: flaming mastiff']

def sbox(text, key):
    out = bytearray()
    for t in text:
        out.append(key[t])
    return bytes(out)

for p in plaintext:
    c = sbox(text, key)
    assert c == sbox(c, ikey), 'Failure'
    print(c)
