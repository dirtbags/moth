#! /usr/bin/python

##
## This is pretty crufty :<
##

import cgitb; cgitb.enable()
import cgi
import os
import fcntl
import re
import sys
from cgi import escape
import Cookie as cookies
from urllib import quote, unquote
from codecs import open
from sets import Set as set
from cStringIO import StringIO

from ctf import pointscli, teams, html, paths

keysfile = os.path.join(paths.LIB, 'puzzler.keys')
datafile = os.path.join(paths.VAR, 'puzzler.dat')
puzzles_dir = os.path.join(paths.WWW, 'puzzler')

##
## This allows you to edit the URL and work on puzzles that haven't been
## unlocked yet.  For now I think that's an okay vulnerability.  It's a
## hacking contest, after all.
##

cat_re = re.compile(r'^[a-z]+$')
points_re = re.compile(r'^[0-9]+$')

points_by_cat = {}
points_by_team = {}
try:
    for line in open(datafile, encoding='utf-8'):
        cat, team, pts = [unquote(v) for v in line.strip().split('\t')]
        pts = int(pts)
        points_by_cat[cat] = max(points_by_cat.get(cat, 0), pts)
        points_by_team.setdefault((team, cat), set()).add(pts)
except IOError:
    pass


c = cookies.SimpleCookie(os.environ.get('HTTP_COOKIE', ''))
try:
    team = c['team'].value
    passwd = c['passwd'].value
except KeyError:
    team, passwd = None, None

f = cgi.FieldStorage()
cat = f.getfirst('c')
points = f.getfirst('p')
team = f.getfirst('t', team)
passwd = f.getfirst('w', passwd)
key = f.getfirst('k', '').decode('utf-8')

def serve(title, sf=None, **kwargs):
    if team or passwd:
        c = cookies.SimpleCookie()
        if team:
            c['team'] = team
        if passwd:
            c['passwd'] = passwd
        print(c)
    if not sf:
        sf = StringIO()
    return html.serve(title, sf.getvalue(), **kwargs)

def safe_join(*args):
    safe = list(args[:1])
    for a in args[1:]:
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

def disabled(cat):
    return os.path.exists(os.path.join(paths.VAR, 'disabled', cat))

def show_cats():
    out = StringIO()
    out.write('<ul>')
    puzzles = os.listdir(puzzles_dir)
    puzzles.sort()
    for p in puzzles:
        if disabled(p):
            continue
        out.write('<li><a href="%spuzzler.cgi?c=%s">%s</a></li>' % (html.base, p, p))
    out.write('</ul>')
    serve('Categories', out)


def show_puzzles(cat, cat_dir):
    out = StringIO()
    opened = points_by_cat.get(cat, 0)
    puzzles = ([int(v) for v in os.listdir(cat_dir)])
    puzzles.sort()
    if puzzles:
        out.write('<ul>')
        for p in puzzles:
            cls = ''
            try:
                if p in points_by_team[(team, cat)]:
                    cls = 'solved'
            except KeyError:
                pass
            out.write('<li><a href="%(base)spuzzler/%(cat)s/%(points)d" class="%(class)s">%(points)d</a></li>' %
                      {'base': html.base,
                       'cat': cat,
                       'points': p,
                       'class': cls})
            if p > opened:
                break
        out.write('</ul>')
    else:
        out.write('<p>None (someone is slacking)</p>')
    serve('Open in %s' % escape(cat), out)

def win(cat, team, points):
    out = StringIO()
    points = int(points)
    f = open(datafile, 'a', encoding='utf-8')
    pointscli.award(cat, team, points)
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.write('%s\t%s\t%d\n' % (quote(cat), quote(team), points))
    out.write('<p>%d points for %s.</p>' % (points, cgi.escape(team)))
    out.write('<p>Back to <a href="%spuzzler.cgi?c=%s">%s</a>.</p>' % (html.base, cat, cat))
    serve('Winner!', out)

def check_key(cat, points, candidate):
    for line in open(keysfile, encoding='utf-8'):
        thiscat, thispoints, key = line.split('\t', 2)
        if (cat, points) == (thiscat, thispoints):
            if  key.rstrip() == candidate:
                return True
    return False

def main():
    cat_dir = safe_join(puzzles_dir, cat)
    points_dir = safe_join(puzzles_dir, cat, points)

    if not cat_dir:
        # Show categories
        show_cats()
    elif not points_dir:
        # Show available puzzles in category
        show_puzzles(cat, cat_dir)
    else:
        if not teams.chkpasswd(team, passwd):
            serve('Wrong password')
        elif not check_key(cat, points, key):
            serve('Wrong key')
        elif int(points) in points_by_team.get((team, cat), set()):
            serve('Greedy greedy')
        else:
            win(cat, team, points)

if __name__ == '__main__':
    import optparse
    import sys, codecs

    sys.stdout = codecs.getwriter('utf-8')(sys.stdout)

    main()


# Local Variables:
# mode: python
# End:
