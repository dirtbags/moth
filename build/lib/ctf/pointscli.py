#! /usr/bin/python

from urllib import quote
import teams
import time
import os
import paths

pointsdir = os.path.join(paths.VAR, 'points')

def award(cat, team, points):
    if not team:
        team = teams.house
    now = time.strftime('%Y-%m-%dT%H:%M:%S')
    pid = os.getpid()
    qcat = quote(cat, '')
    qteam = quote(team, '')
    basename = '%s.%d.%s.%s' % (now, pid, qcat, qteam)
    # FAT can't handle :
    basename = basename.replace(':', '.')
    tmpfn = os.path.join(pointsdir, 'tmp', basename)
    curfn = os.path.join(pointsdir, 'cur', basename)
    f = open(tmpfn, 'w')
    f.write('%s\t%s\t%s\t%d\n' % (now, cat, team, points))
    f.close()
    os.rename(tmpfn, curfn)

def main():
    import optparse

    p = optparse.OptionParser('%prog CATEGORY TEAM POINTS')
    opts, args = p.parse_args()
    if len(args) != 3:
        p.error('Wrong number of arguments')
    cat, team, points = args
    points = int(points)
    award(cat, team, points)

if __name__ == '__main__':
    main()
