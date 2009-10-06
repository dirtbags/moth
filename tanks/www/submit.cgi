#!/usr/bin/python

import cgi
import cgitb; cgitb.enable()
import os

try:
    from urllib.parse import quote
except:
    from urllib import quote

try:
    from ctf import teams
except:
    import sys
    path = '/home/pflarr/repos/gctf/'
    sys.path.append(path)
    from ctf import teams

print """Content-Type: text/html\n\n"""
print """<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Strict//EN">\n\n"""
head = open('head.html').read() % "Submission Results"
print head
print "<H1>Results</H1>"
print '<a href="submit.html">Submit</a> | <a href="results.cgi">Results</a> | <a href="docs.cgi">Documentation</a>'

def done():
    print '</body></html>'
    sys.exit(0)

fields = cgi.FieldStorage()
team = fields.getfirst('team', '').strip()
passwd = fields.getfirst('passwd', '').strip()
code = fields.getfirst('code', '')
if not team:
    print '<p>No team specified'; done()
elif not passwd:
    print '<p>No password given'; done()
elif not code:
    print '<p>No program given.'; done()

if team not in teams.teams:
    print '<p>Team is not registered.'; done()

if passwd != teams.teams[team][0]:
    print '<p>Invalid password.'; done()

path = os.path.join('data/ai/players', encode(team) )
file = open(path, 'w')
file.write(code)
file.close()

done()
