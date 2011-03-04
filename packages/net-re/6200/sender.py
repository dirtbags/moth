from socket import *
import random
import sys

data = open('twain.txt', 'rb').read()
sbox_file = open('sbox.txt', 'wb')

try:
    src_port = int(sys.argv[1])
    dest_ip = sys.argv[2]
    dest_port = int(sys.argv[3])
except:
    print "Usage: python sender.py src_port dest_ip dest_port"

random.seed(1)
sbox = []
l = range(256)
for i in range(256):
    v = random.choice(l)
    sbox.append(v)
    l.remove(v)
    sbox_file.write('%02x' % v)
sbox_file.close()

data = [ord(c) for c in data]
ciphertext = []
for c in data:
    row = (c & 0xf0) >> 4
    col = c & 0x0f
    index = row * 16 + col
    ciphertext.append(chr(sbox[index]))
ciphertext = ''.join(ciphertext)

sock = socket(AF_INET, SOCK_STREAM)
addr = ("", src_port)
sock.bind((addr))
sock.connect((dest_ip, dest_port))
sock.send(ciphertext)
