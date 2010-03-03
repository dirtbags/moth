#! /usr/bin/python

import os
from ctf import html, paths
from cgi import escape

basedir = os.path.join(paths.VAR, 'tanks')

links = '''
      <h3>Tanks</h3>
      <li><a href="docs.html">Docs</a></li>
      <li><a href="results.cgi">Results</a></li>
      <li><a href="submit.html">Submit</a></li>
      <li><a href="errors.cgi">My Errors</a></li>
'''

body = []

body.append('<h1>Last Winner:</h1>')
body.append('<p>')
body.append(escape(open(os.path.join(basedir, 'winner')).read()))
body.append('</p>')
body.append('<h1>Results so far:</h1>')
body.append('<ul>')
results = os.listdir(os.path.join(basedir, 'results'))
results.sort()
results.reverse()
for fn in results:
    num, _ = os.path.splitext(fn)
    body.append('<li><a href="results/%s">%s</a></li>' % (fn, num))
body.append('</ul>')

html.serve('Tanks Results', '\n'.join(body), links=links)
