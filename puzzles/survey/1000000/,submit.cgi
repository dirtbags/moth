#! /usr/bin/env python3

import cgi
import time
import os

f = cgi.FieldStorage()
if f.getfirst('submit'):
    print('Content-type: text/plain')
    print()
    print('Thanks for filling in the survey.')
    print()
    try:
        fn = '/var/lib/ctf/survey/%s.%d.%d.txt' % (time.strftime('%Y-%m-%d'), time.time(), os.getpid())
        o = open(fn, 'w')
        for k in f.keys():
            o.write('%s: %r\n' % (k, f.getlist(k)))
    except IOError:
        pass
    print('The key is:')
    print('    quux blorb frotz')
else:
    print('Content-type: text/plain')
    print()
    print('You need to actually fill in the form to get the key.')
