#! /usr/bin/env python3

import sys
import random

primes = [2, 3, 5, 7, 11, 13, 17, 19]
letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'

data = sys.stdin.read().strip()
jumble = ''.join(data.split())

lj = len(jumble)
below = (0, 0)
above = (lj, 2)
for i in primes:
    for j in primes:
        m = i * j
        if (m < lj) and (m > below[0] * below[1]):
            below = (i, j)
        elif (m >= lj) and (m < (above[0] * above[1])):
            above = (i, j)

for i in range(lj, (above[0] * above[1])):
    jumble += random.choice(letters)

out = []
for i in range(above[0]):
    for j in range(above[1]):
        out.append(jumble[j*above[0] + i])
print(''.join(out))
