msg = 'hello crypto!'

v = [0x3a, 0x21]
k = (0x5A, 0xe2)
ct = []
for i in range(0, len(msg), len(k)):
    for j in range(len(k)):
        if (i+j < len(msg)):
            p = ord(msg[i+j])
            r = (p ^ v[j]) ^ k[j]
            v[j] = r
            ct.append(r)

for v in ct:
    print '%02x' % v,
print 

