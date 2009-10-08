#!/usr/bin/python3

print("""Content-Type: text/html\n\n""")
print("""<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Strict//EN">\n\n""")
import cgi
import cgitb; cgitb.enable()
import sys
import os

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
teams.build_teams()

head = open('head.html').read() % "Error Report"
print(head)
print('<H1>Your Errors</H1>')
print(open('links.html').read())

def done():
    print('</body></html>')
    sys.exit(0)

fields = cgi.FieldStorage()
team = fields.getfirst('team', '').strip()
passwd = fields.getfirst('passwd', '').strip()
if team and passwd and \
   team in teams.teams and passwd == teams.teams[team][0]:
    path = os.path.join(Config.DATA_PATH, 'errors', quote(team))
    if os.path.isfile(path):
        errors = open(path).readlines()
        print('<p>Your latest errors:')
        print('<div class=errors>')
        if errors:
            print('<BR>\n'.join(errors))
        else:
            print('There were no errors.')
        print('</div>')
    else:
        print('<p>No error file found.')

    done()

if team and team not in teams.teams:
    print('<p>Invalid team.')

if team and team in teams.teams and passwd != teams.teams[team][0]:
    print('<p>Invalid password.')

print('''
<form action="errors.cgi" method="get">
    <fieldset>
        <legend>Error report request:</legend>
        Team: <input type="text" name="team"><BR>
        Password: <input type="text" name="passwd"><BR>
        <button type="get my errors">Submit</button>
    </fieldset>
</form>''')

done()
