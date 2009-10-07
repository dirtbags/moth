#! /usr/bin/env python3

import cgi
import os
import fcntl
import re
import sys
import http.cookies
from urllib.parse import quote, unquote
from . import config
from . import pointscli
from . import teams

datafile = config.datafile('puzzler.dat')
keysfile = config.get('puzzler', 'keys_file')
puzzles_dir = config.get('puzzler', 'dir')
cgi_url = config.get('puzzler', 'cgi_url')
base_url = config.get('puzzler', 'base_url')

##
## This allows you to edit the URL and work on puzzles that haven't been
## unlocked yet.  For now I think that's an okay vulnerability.  It's a
## hacking contest, after all.
##

cat_re = re.compile(r'^[a-z]+$')
points_re = re.compile(r'^[0-9]+$')

def dbg(*vals):
    print('<!--: \nContent-type: text/html\n\n--><pre>')
    print(*vals)
    print('</pre>')


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


c = http.cookies.SimpleCookie(os.environ.get('HTTP_COOKIE', ''))
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
key = f.getfirst('k')

def start_html(title):
    if team or passwd:
        c = http.cookies.SimpleCookie()
        if team:
            c['team'] = team
        if passwd:
            c['passwd'] = passwd
        print(c)
    config.start_html(title)

end_html = config.end_html

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

def show_cats():
    start_html('Categories')
    print('<ul>')
    for p in sorted(os.listdir(puzzles_dir)):
        if config.disabled(p):
            continue
        print('<li><a href="%s?c=%s">%s</a></li>' % (cgi_url, p, p))
    print('</ul>')
    end_html()


def show_puzzles(cat, cat_dir):
    start_html('Open in %s' % cat)
    opened = points_by_cat.get(cat, 0)
    puzzles = sorted([int(v) for v in os.listdir(cat_dir)])
    if puzzles:
        print('<ul>')
        for p in puzzles:
            cls = ''
            try:
                if p in points_by_team[(team, cat)]:
                    cls = 'solved'
            except KeyError:
                pass
            print('<li><a href="%(base)s/%(cat)s/%(points)d" class="%(class)s">%(points)d</a></li>' %
                  {'base': base_url,
                   'cat': cat,
                   'points': p,
                   'class': cls})
            if p > opened:
                break
        print('</ul>')
    else:
        print('<p>None (someone is slacking)</p>')
    end_html()

def win(cat, team, points):
    start_html('Winner!')
    points = int(points)
    f = open(datafile, 'a', encoding='utf-8')
    pointscli.submit(cat, team, points)
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.write('%s\t%s\t%d\n' % (quote(cat), quote(team), points))
    print('<p>%d points for %s.</p>' % (points, team))
    print('<p>Back to <a href="%s?c=%s">%s</a>.</p>' % (cgi_url, cat, cat))
    end_html()

def get_key(cat, points):
    for line in open(keysfile, encoding='utf-8'):
        thiscat, thispoints, ret = line.split('\t', 2)
        if (cat, points) == (thiscat, thispoints):
            return ret.strip()
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
        thekey = get_key(cat, points)
        if not teams.chkpasswd(team, passwd):
            start_html('Wrong password')
            end_html()
        elif key != thekey:
            start_html('Wrong key')
            end_html()
        elif int(points) in points_by_team.get((team, cat), set()):
            start_html('Greedy greedy')
            end_html()
        else:
            win(cat, team, points)

if __name__ == '__main__':
    import optparse

    parser = optparse.OptionParser('%prog CATEGORY POINTS')
    opts, args = parser.parse_args()

    if len(args) == 2:
        cat, points = args
        show_puzzle(cat, points)
    else:
        parser.print_usage()


# Local Variables:
# mode: python
# End:
