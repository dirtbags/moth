#! /usr/bin/env python3

import time
import os
import tempfile
from . import points
from . import teams
from . import config

pngout = config.datafile('histogram.png')

def main(s=None):
    scores = {}
    now = 0

    if not s:
        s = points.Storage()

    plotparts = []

    catscores = {}
    for cat in s.categories():
        catscores[cat] = s.cat_points(cat)

    scoresfile = tempfile.NamedTemporaryFile('w')
    fn = scoresfile.name
    i = 2
    for team in s.teams:
        plotparts.append('"%s" using 1:%d with lines linewidth 2 linetype rgb "#%s"' % (fn, i, teams.color(team)))
        scores[team] = 0
        i += 1

    def write_scores(t):
        scoresfile.write('%d' % t)
        for team in s.teams:
            scoresfile.write('\t%f' % (scores[team]))
        scoresfile.write('\n')

    for when, cat, team, score in s.log:
        if when > now:
            if now:
                write_scores(now)
            now = when
        pct = score / catscores[cat]
        scores[team] += pct
        #print('%d [%s] [%s] %d' % (when, cat, team, points))

    write_scores(when)
    scoresfile.flush()

    instructions = tempfile.NamedTemporaryFile('w')
    instructions.write('''
set style data lines
set xdata time
set timefmt "%%s"
set format ""
set border 3
set xtics nomirror
set ytics nomirror
set nokey
set terminal png transparent size 640,200 x000000 xffffff
set output "%(pngout)s"
plot %(plot)s\n''' % {'plot': ','.join(plotparts),
                      'pngout': pngout})
    instructions.flush()

    gp = os.system('gnuplot %s 2>/dev/null </dev/null' % instructions.name)

if __name__ == '__main__':
    main()
