#! /usr/bin/env python3

import fcntl
import time
import os
from urllib.parse import quote, unquote

house = 'dirtbags'

passwdfn = '/var/lib/ctf/passwd'

teams = None
built = 0
def build_teams():
    global teams, built

    modt = os.path.getmtime(passwdfn)
    if modt <= built:
        return

    teams = {}
    try:
        f = open(passwdfn)
        for line in f:
            line = line.strip()
            team, passwd = [unquote(v) for v in line.strip().split('\t')]
            teams[team] = passwd
    except IOError:
        pass
    built = time.time()

def validate(team):
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
