#! /usr/bin/env python3

import points
import time
import os
import tempfile

def main(s=None):
    scores = {}
    now = 0

    if not s:
        s = points.Storage('scores.dat')

    plotparts = []
    teams = s.teams()
    teamcolors = points.colors(teams)

    catscores = {}
    for cat in s.categories():
        catscores[cat] = s.cat_points(cat)

    scoresfile = tempfile.NamedTemporaryFile('w')
    fn = scoresfile.name
    i = 2
    for team in teams:
        plotparts.append('"%s" using 1:%d with lines linewidth 2 linetype rgb "#%s"' % (fn, i, teamcolors[team]))
        scores[team] = 0
        i += 1

    def write_scores(t):
        scoresfile.write('%d' % t)
        for team in teams:
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

    gp = os.popen('gnuplot 2> /dev/null', 'w')
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
    gp.close()

if __name__ == '__main__':
    main()
