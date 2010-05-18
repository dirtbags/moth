#! /usr/bin/python

import os
import string
import sys
from codecs import open

from paths import *

template_fn = os.path.join(LIB, 'template.html')
template = string.Template(open(template_fn, encoding='utf-8').read())

base = BASE_URL
css = base + 'ctf.css'

def substitute(title, body, base=base, hdr='', body_class='', onload='', links=''):
    return template.substitute(title=title,
                               hdr=hdr,
                               body_class=body_class,
                               base=base,
                               links=links,
                               body=body)

def serve(title, body, **kwargs):
    out = substitute(title, body, **kwargs)
    print 'Content-type: text/html'
    print 'Content-length: %d' % len(out)
    print
    sys.stdout.write(out)
    sys.stdout.flush()

def write(filename, title, body, **kwargs):
    f = open(filename, 'w', encoding='utf-8')
    f.write(substitute(title, body, **kwargs))
