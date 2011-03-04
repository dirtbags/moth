from socket import *
from hashlib import md5
import sys

plaintext = open('clever_girl.jpg', 'rb').read()

try:
    src_port = int(sys.argv[1])
    dest_ip = sys.argv[2]
    dest_port = int(sys.argv[3])
except:
    print "Usage: python sender.py src_port dest_ip dest_port"
    sys.exit(1)

key = 'why am I the key'
data = [ord(c) for c in plaintext]
ciphertext = [key]
hasher = md5(key)
for c in data:
    digest = hasher.digest()
    hasher.update(digest)
    ciphertext.append(chr(c ^ ord(digest[0])))
ciphertext = ''.join(ciphertext)
print ciphertext[:16]

key = ciphertext[:16]
hasher = md5(key)
decrypted = []
for c in ciphertext[16:]:
    digest = hasher.digest()
    hasher.update(digest)
    decrypted.append(chr(ord(c) ^ ord(digest[0])))
decrypted = ''.join(decrypted)

assert decrypted == plaintext

sock = socket(AF_INET, SOCK_STREAM)
addr = ("", src_port)
sock.bind((addr))
sock.connect((dest_ip, dest_port))
sock.send(ciphertext)
