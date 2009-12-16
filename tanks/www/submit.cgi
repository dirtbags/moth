#!/usr/bin/python3

import cgi
import cgitb; cgitb.enable()
import os
import sys

import Config

try:
    from urllib.parse import quote
except:
    from urllib import quote

try:
    from ctf import teams
except:
    path = '/home/pflarr/repos/gctf/'
    sys.path.append(path)
    from ctf import teams
from ctf import config
teams.build_teams()

print(config.start_html('Tanks Submission',
                        links_title='Tanks',
                        links=[('docs.cgi', 'Docs'),
                               ('results.cgi', 'Results'),
                               ('submit.html', 'Submit'),
                               ('errors.cgi', 'My Errors')]))

def done():
    print(config.end_html())
    sys.exit(0)

fields = cgi.FieldStorage()
team = fields.getfirst('team', '').strip()
passwd = fields.getfirst('passwd', '').strip()
code = fields.getfirst('code', '')
if not team:
    print('<p>No team specified</p>'); done()
elif not passwd:
    print('<p>No password given</p>'); done()
elif not code:
    print('<p>No program given.</p>'); done()

if team not in teams.teams:
    print('<p>Team is not registered.</p>'); done()

if passwd != teams.teams[team][0]:
    print('<p>Invalid password.</p>'); done()

path = os.path.join(Config.DATA_PATH, 'ai/players', quote(team, safe='') )
file = open(path, 'w')
file.write(code)
file.close()

print("<p>Submission successful.</p>")

done()
