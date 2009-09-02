#! /usr/bin/env python3

import fcntl
from urllib.parse import quote, unquote

house = 'dirtbags'

teams = None
def build_teams():
    global teams

    teams = {}
    try:
        f = open('passwd')
        for line in f:
            line = line.strip()
            team, passwd = [unquote(v) for v in line.strip().split('\t')]
            teams[team] = passwd
    except IOError:
        pass

def validate(team):
    if teams is None:
        build_teams()

def chkpasswd(team, passwd):
    validate(team)
    if teams.get(team) == passwd:
        return True
    else:
        return False

def exists(team):
    validate(team)
    if team == house:
        return True
    return team in teams

def add(team, passwd):
    f = open('passwd', 'a')
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.seek(0, 2)
    f.write('%s\t%s\n' % (quote(team), quote(passwd)))
