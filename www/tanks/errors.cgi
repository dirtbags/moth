#!/usr/bin/python

import cgi
import cgitb; cgitb.enable()
import sys
import os

from urllib import quote

from ctf import teams, html

basedir = '/var/lib/ctf/tanks'

links = '''
      <h3>Tanks</h3>
      <li><a href="docs.html">Docs</a></li>
      <li><a href="results.cgi">Results</a></li>
      <li><a href="submit.html">Submit</a></li>
      <li><a href="errors.cgi">My Errors</a></li>
'''

body = []
fields = cgi.FieldStorage()
team = fields.getfirst('team', '').strip()
passwd = fields.getfirst('passwd', '').strip()
if not team:
    pass
elif teams.chkpasswd(team, passwd):
    path = os.path.join(basedir, 'errors', quote(team))
    if os.path.isfile(path):
        body.append('<p>Your latest errors:</p>')
        errors = open(path).readlines()
        if errors:
            body.append('<ul class="errors">')
            for e in errors:
                body.append('<li>%s</li>' % cgi.escape(e))
            body.append('</ul>')
        else:
            body.append('<p>There were no errors.</p>')
    else:
        body.append('<p>No error file found.</p>')
else:
    body.append('Authentication failed.')

body.append('''
<form action="errors.cgi" method="get">
    <fieldset>
        <legend>Error report request:</legend>
        Team: <input type="text" name="team"/><br/>
        Password: <input type="password" name="passwd"/><br/>
        <button type="get my errors">Submit</button>
    </fieldset>
</form>''')

html.serve('Tanks Errors', '\n'.join(body), links=links)

