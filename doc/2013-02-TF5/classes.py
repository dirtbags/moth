#! /usr/bin/python3

import sys

limits = {
	"net": 100,
	"mal": 60,
	"hst": 60,
	"icc": 40,
	"nil": 9000,
	":-(": 9000,
}

# Read in allowed substring list
allowed = []
for line in open("approved.txt"):
	allowed.append(line.strip())

# Read in registration data
registrants = []
regs = {}
for line in open("reg.txt"):
	line = line.strip('\n')
	ok = False
	for a in allowed:
		if a in line:
			ok = True
			break
	name, email, org, c1, c2, _ = line.split('\t')
	if not ok:
		c1 = c2 = "nil"

	if email not in registrants:
		registrants.append(email)
	regs[email] = (name, org, c1, c2)

# Divvy out classes
which = {}
counts = [0, 0, 0]
for email in registrants:
	r = regs.get(email)
	name, org, c1, c2 = regs[email]
	c = [c1, c2, ":-("]

	for i in range(3):
		w = c[i]
		if limits[w] > 0:
			which[email] = i
			limits[w] -= 1
			counts[i] += 1
			break
	
	try:
		oldreg = open(email).read()
	except:
		oldreg = None
		
	if oldreg != w:
		print(w, email)
		open(email, "w").write(w)

print("%d got 1st choice, %d got 2nd choice, %d got screwed" %
	(counts[0], counts[1], counts[2]))
