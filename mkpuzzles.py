#! /usr/bin/env python3

import os
import shutil
import optparse
import config

p = optparse.OptionParser()
p.add_option('-p', '--puzzles', dest='puzzles', default='puzzles',
             help='Directory containing puzzles')
p.add_option('-w', '--htmldir', dest='htmldir', default='puzzler',
             help='Directory to write HTML puzzle tree')
p.add_option('-k', '--keyfile', dest='keyfile', default='puzzler.keys',
             help='Where to write keys')

opts, args = p.parse_args()

keys = []

for cat in os.listdir(opts.puzzles):
    dirname = os.path.join(opts.puzzles, cat)
    for points in os.listdir(dirname):
        pointsdir = os.path.join(dirname, points)
        outdir = os.path.join(opts.htmldir, cat, points)
        try:
            os.makedirs(outdir)
        except OSError:
            pass

        readme = ''
        files = []
        for fn in os.listdir(pointsdir):
            path = os.path.join(pointsdir, fn)
            if fn == 'key':
                key = open(path, encoding='utf-8').readline().strip()
                keys.append((cat, points, key))
            elif fn == 'index.html':
                readme = open(path, encoding='utf-8').read()
            elif fn.endswith('~'):
                pass
            else:
                files.append((fn, path))

        title = '%s for %s points' % (cat, points)
        f = open(os.path.join(outdir, 'index.html'), 'w', encoding='utf-8')
        f.write('''<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC
  "-//W3C//DTD XHTML 1.0 Strict//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title>%(title)s</title>
    <link rel="stylesheet" href="%(css)s" type="text/css" />
  </head>
  <body>
    <h1>%(title)s</h1>
''' % {'title': title,
       'css': config.css})
        if readme:
            f.write('<div class="readme">%s</div>\n' % readme)
        if files:
            f.write('<ul>\n')
            for fn, path in files:
                shutil.copy(path, outdir)
                f.write('<li><a href="%s">%s</a></li>\n' % (fn, fn))
            f.write('</ul>\n')
        f.write('''
    <form action="%(cgi)s" method="post">
      <fieldset>
        <legend>Your answer:</legend>
        <input type="hidden" name="c" value="%(cat)s" />
        <input type="hidden" name="p" value="%(points)s" />
        Team: <input name="t" /><br />
        Password: <input type="password" name="w" /><br />
        Key: <input name="k" /><br />
        <input type="submit" />
      </fieldset>
    </form>
  </body>
</html>
''' % {'cgi': config.get('puzzler', 'cgi_url'),
       'cat': cat,
       'points': points})

f = open(opts.keyfile, 'w', encoding='utf-8')
for key in keys:
    f.write('%s\t%s\t%s\n' % key)

