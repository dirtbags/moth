#! /usr/bin/env python3

import cgitb; cgitb.enable()
import cgi
import os
import fcntl
import re
import sys
import pointscli
import teams
from urllib.parse import quote, unquote

##
## This allows you to edit the URL and work on puzzles that haven't been
## unlocked yet.  For now I think that's an okay vulnerability.  It's a
## hacking contest, after all.
##

cat_re = re.compile(r'^[a-z]+$')
points_re = re.compile(r'^[0-9]+$')

def dbg(*vals):
    print('Content-type: text/plain\n\n')
    print(*vals)


points_by_cat = {}
points_by_team = {}
try:
    for line in open('puzzler.dat'):
        cat, team, pts = [unquote(v) for v in line.strip().split('\t')]
        pts = int(pts)
        points_by_cat[cat] = max(points_by_cat.get(cat, 0), pts)
        points_by_team.setdefault((team, cat), set()).add(pts)
except IOError:
    pass


f = cgi.FieldStorage()

cat = f.getfirst('c')
points = f.getfirst('p')
team = f.getfirst('t')
passwd = f.getfirst('w')
key = f.getfirst('k')

verboten = ['key', 'index.html']

def start_html(title):
    print('''Content-type: text/html

<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC
  "-//W3C//DTD XHTML 1.0 Strict//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title>%s</title>
    <link rel="stylesheet" href="ctf.css" type="text/css" />
  </head>
  <body>
    <h1>%s</h1>
''' % (title, title))

def end_html():
    print('</body></html>')


def safe_join(*args):
    safe = []
    for a in args:
        if not a:
            return None
        else:
            a = a.replace('..', '')
            a = a.replace('/', '')
        safe.append(a)
    ret = '/'.join(safe)
    if os.path.exists(ret):
        return ret

def dump_file(fn):
    f = open(fn, 'rb')
    while True:
        d = f.read(4096)
        if not d:
            break
        sys.stdout.buffer.write(d)

def show_cats():
    start_html('Categories')
    print('<ul>')
    for p in sorted(os.listdir('puzzles')):
        print('<li><a href="puzzler.cgi?c=%s">%s</a></li>' % (p, p))
    print('</ul>')
    end_html()


def show_puzzles(cat, cat_dir):
    start_html('Open in %s' % cat)
    opened = points_by_cat.get(cat, 0)
    puzzles = sorted([int(v) for v in os.listdir(cat_dir)])
    if puzzles:
        print('<ul>')
        for p in puzzles:
            print('<li><a href="puzzler.cgi?c=%s&p=%d">%d</a></li>' % (cat, p, p))
            if p > opened:
                break
        print('</ul>')
    else:
        print('<p>None (someone is slacking)</p>')
    end_html()

def show_puzzle(cat, points, points_dir, team, passwd):
    # Show puzzle in cat for points
    start_html('%s for %s' % (cat, points))
    fn = os.path.join(points_dir, 'index.html')
    if os.path.exists(fn):
        print('<div class="readme">')
        dump_file(fn)
        print('</div>')
    print('<ul>')
    for fn in sorted(os.listdir(points_dir)):
        if fn.endswith('~') or fn.startswith('.') or fn in verboten:
            continue
        print('<li><a href="puzzler.cgi?c=%s&p=%s&f=%s">%s</a></li>' % (cat, points, fn, fn))
    print('</ul>')
    print('<form action="puzzler.cgi" method="post">')
    print('<input type="hidden" name="c" value="%s" />' % cat)
    print('<input type="hidden" name="p" value="%s" />' % points)
    print('Team: <input name="t" value="%s" /><br />' % (team or ''))
    print('Password: <input type="password" name="w" value="%s" /><br />' % (passwd or ''))
    print('Key: <input name="k" /><br />')
    print('<input type="submit" />')
    print('</form>')
    end_html()

def win(cat, team, points):
    start_html('Winner!')
    points = int(points)
    f = open('puzzler.dat', 'a')
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.write('%s\t%s\t%d\n' % (quote(cat), quote(team), points))
    pointscli.submit(cat, team, points)
    print('<p>%d points for %s.</p>' % (team, points))
    end_html()

def main():
    cat_dir = safe_join('puzzles', cat)
    points_dir = safe_join('puzzles', cat, points)

    if not cat_dir:
        # Show categories
        show_cats()
    elif not points_dir:
        # Show available puzzles in category
        show_puzzles(cat, cat_dir)
    elif not (team and passwd and key):
        fn = f.getfirst('f')
        if fn in verboten:
            fn = None
        fn = safe_join('puzzles', cat, points, fn)
        if fn:
            # Provide a file from this directory
            print('Content-type: application/octet-stream')
            print()
            dump_file(fn)
        else:
            show_puzzle(cat, points, points_dir, team, passwd)
    else:
        thekey = open('%s/key' % points_dir).read().strip()
        if not teams.chkpasswd(team, passwd):
            start_html('Wrong password')
            end_html()
        elif key != thekey:
            show_puzzle(cat, points, points_dir)
        elif points_by_team.get((team, cat)):
            start_html('Greedy greedy')
            end_html()
        else:
            win(cat, team, points)

main()
