#! /usr/bin/python3

import smtplib
import sys

smtpd = smtplib.SMTP("mail.lanl.gov")

template = open("netre-email.txt").read()
assert 'RCPT' in template
assert 'TOKEN' in template

for line in open("netarch-tokens.txt"):
	email, token = line.strip().split()

	print(email)
	msg = template.replace("RCPT", email).replace("TOKEN", token)
	smtpd.sendmail("neale@lanl.gov", [email], msg)
	#print(msg)
