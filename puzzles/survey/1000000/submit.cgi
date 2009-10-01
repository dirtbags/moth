#! /usr/bin/env python3

import cgi

f = cgi.FieldStorage()
if f.getfirst('submit'):
    print('Content-type: text/plain')
    print()
    print('Thanks for filling in the survey.')
    print()
    print('The key is:')
    print('    quux blorb frotz')
else:
    print('Content-type: text/plain')
    print()
    print('You need to actually fill in the form to get the key.')
