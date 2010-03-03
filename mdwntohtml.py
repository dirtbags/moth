#! /usr/bin/python

import os
import shutil
import optparse
import string
import markdown
from codecs import open

p = optparse.OptionParser('%prog [OPTIONS] infile outfile')
p.add_option('-t', '--template', dest='template', default='template.html',
             help='Location of HTML template')
p.add_option('-b', '--base', dest='base', default='',
             help='Base URL for contest')

opts, args = p.parse_args()

basedir = os.path.dirname(args[0])
links_fn = os.path.join(basedir, 'links.xml')
try:
    links = open(links_fn, encoding='utf-8').read()
except IOError:
    links = ''

f = open(args[0], encoding='utf-8')
title = ''
for line in f:
    line = line.strip()
    if not line:
        break
    k, v = line.split(': ')
    if k.lower() == 'title':
        title = v
body = markdown.markdown(f.read(99999))
template = string.Template(open(opts.template, encoding='utf-8').read())
page = template.substitute(hdr='',
                           title=title,
                           base=opts.base,
                           links=links,
                           body_class='',
                           onload='',
                           body=body)

open(args[1], 'w', encoding='utf-8').write(page)
