#! /usr/bin/env python3

import points
import time
import os

teamfiles = {}
scores = {}
now = 0

s = points.Storage('scores.dat')

plotparts = []
teams = s.teams()
teamcolors = points.colors(teams)

fn = 'scores.hist'
scoresfile = open(fn, 'w')
i = 2
for team in teams:
    plotparts.append('"%s" using 1:%d with lines linewidth 2 linetype rgb "#%s"' % (fn, i, teamcolors[team]))
    scores[team] = 0
    i += 1

def write_scores(t):
    scoresfile.write('%d' % t)
    for team in teams:
        scoresfile.write('\t%d' % (scores[team]))
    scoresfile.write('\n')

for when, cat, team, points in s.log:
    if when > now:
        if now:
            write_scores(now)
        now = when
    scores[team] += points
    #print('%d [%s] [%s] %d' % (when, cat, team, points))

write_scores(when)

for f in teamfiles.values():
    f.close()

gp = os.popen('gnuplot > /dev/null', 'w')
gp.write('set style data lines\n')
gp.write('set xdata time\n')
gp.write('set timefmt "%s"\n')
gp.write('set format ""\n')
gp.write('set border 3\n')
gp.write('set xtics nomirror\n')
gp.write('set ytics nomirror\n')
gp.write('set nokey\n')
gp.write('set terminal png transparent size 640,200 x000000 xffffff\n')
gp.write('set output "histogram.png"\n')
gp.write('plot %s\n' % ','.join(plotparts))
gp.flush()
