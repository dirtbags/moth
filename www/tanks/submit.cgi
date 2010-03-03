#!/usr/bin/python

import cgi
import cgitb; cgitb.enable()
import os
import sys

from urllib import quote

from ctf import teams, html, paths

basedir = os.path.join(paths.VAR, 'tanks')

links = '''
      <h3>Tanks</h3>
      <li><a href="docs.html">Docs</a></li>
      <li><a href="results.cgi">Results</a></li>
      <li><a href="submit.html">Submit</a></li>
      <li><a href="errors.cgi">My Errors</a></li>
'''


fields = cgi.FieldStorage()
team = fields.getfirst('team', '').strip()
passwd = fields.getfirst('passwd', '').strip()
code = fields.getfirst('code', '')
if not teams.chkpasswd(team, passwd):
    body = '<p>Authentication failed.</p>'
elif not code:
    body = '<p>No program given.</p>'
else:
    path = os.path.join(basedir, 'ai/players', quote(team, safe=''))
    file = open(path, 'w')
    file.write(code)
    file.close()

    body = ("<p>Submission successful.</p>")

html.serve('Tanks Submission', body, links=links)
