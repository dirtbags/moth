#! /usr/bin/python3

import smtplib

smtpd = smtplib.SMTP("mail.lanl.gov")

courses = {
	"net": "Network Archaeology",
	"mal": "Malware Reverse-Engineering",
	"hst": "Host Forensics",
	"icc": "Incident Coordination",
	"nil": "None",
}

limits = {
	"net": 120,
	"mal": 70,
	"hst": 70,
	"icc": 40,
	"nil": 9000,
	":-(": 9000,
}

# Read in allowed substring list
allowed = []
for line in open("approved.txt"):
	line = line.strip()
	if line:
		allowed.append(line)
template = open("mail.txt").read()

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
		#msg = template.replace("RCPT", email).replace("COURSE", courses[w])
		#smtpd.sendmail("neale@lanl.gov", [email], msg)
		#open(email, "w").write(w)

	#print(w, email, r)


print("%d got 1st choice, %d got 2nd choice, %d got screwed" % (counts[0], counts[1], counts[2]))
