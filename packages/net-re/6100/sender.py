from socket import *

data = open('exe', 'rb').read()

bs = 1020
pos = 0
tid = 0x52ab
pid = 0

key = 0x52
IV = 0xe3

data = [ord(c) for c in data]
ciphertext = []
for c in data:
    ct = (IV ^ c) ^ key
    IV = ct
    ciphertext.append(ct)
ciphertext = ''.join([chr(c) for c in ciphertext])


sock = socket(AF_INET, SOCK_DGRAM)
addr = ("127.0.0.1", 49143)

while (pos < len(data)):
    sock.sendto("%04x%04x%s" % (tid, pid, data[pos:pos+bs]), addr)
    pos = pos + bs
    pid += 1
