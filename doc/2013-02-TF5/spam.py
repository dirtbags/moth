#! /usr/bin/python3

import smtplib
import sys

courses = {
	"net": "Network Archaeology",
	"mal": "Malware Reverse-Engineering",
	"hst": "Host Forensics",
	"icc": "Incident Coordination",
	"nil": "None",
}

smtpd = smtplib.SMTP("mail.lanl.gov")

template = sys.stdin.read()
if 'RCPT' not in template:
	print("Pass the template on stdin.")
else:
	for line in open("assignments.txt"):
		course, email = line.strip().split()
		coursename = courses[course]
	
		print(email)
		msg = template.replace("RCPT", email).replace("COURSE", coursename)
		smtpd.sendmail("neale@lanl.gov", [email], msg)
		#print(msg)
