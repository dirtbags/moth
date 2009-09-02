#! /usr/bin/env python3

import fcntl

house = 'dirtbags'

teams = None

def build_teams():
    global teams

    teams = {}
    try:
        f = open('passwd')
        for line in f:
            team, passwd = line.strip().split('\t')
            teams[team] = passwd
    except IOError:
        pass

def chkpasswd(team, passwd):
    if teams is None:
        build_teams()
    if teams.get(team) == passwd:
        return True
    else:
        return False

def exists(team):
    if teams is None:
        build_teams()
    if team == house:
        return True
    return team in teams

def add(team, passwd):
    f = open('passwd', 'a')
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.seek(0, 2)
    f.write('%s\t%s\n' % (team, passwd))
