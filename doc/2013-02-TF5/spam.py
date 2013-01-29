#! /usr/bin/python3

import smtplib

smtpd = smtplib.SMTP("mail.lanl.gov")

template = open("hubs-mail.txt").read()

for line in open("assignments.txt"):
	course, email = line.strip().split()

	print(email)
	msg = template.replace("RCPT", email)
	smtpd.sendmail("neale@lanl.gov", [email], msg)
