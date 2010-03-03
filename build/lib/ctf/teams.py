#! /usr/bin/python

import fcntl
import time
import os
from urllib import quote, unquote
import paths

house = 'dirtbags'
passwdfn = os.path.join(paths.VAR, 'passwd')
team_colors = ['F0888A', '88BDF0', '00782B', '999900', 'EF9C00',
               'F4B5B7', 'E2EFFB', '89CA9D', 'FAF519', 'FFE7BB',
               'BA88F0', '8DCFF4', 'BEDFC4', 'FFFAB2', 'D7D7D7',
               'C5B9D7', '006189', '8DCB41', 'FFCC00', '898989']

teams = {}
built = 0
def build_teams():
    global teams, built
    if not os.path.exists(passwdfn):
        return
    if os.path.getmtime(passwdfn) <= built:
        return

    teams = {}
    try:
        f = open(passwdfn)
        for line in f:
            line = line.strip()
            if not line:
                continue
            team, passwd, color = map(unquote, line.strip().split('\t'))
            teams[team] = (passwd, color)
    except IOError:
        pass
    built = time.time()

def validate(team):
    build_teams()

def chkpasswd(team, passwd):
    validate(team)
    if teams.get(team, [None, None])[0] == passwd:
        return True
    else:
        return False

def exists(team):
    validate(team)
    if team == house:
        return True
    return team in teams

def add(team, passwd):
    build_teams()
    color = team_colors[len(teams)%len(team_colors)]

    assert team not in teams, "Team already exists."

    f = open(passwdfn, 'a')
    fcntl.lockf(f, fcntl.LOCK_EX)
    f.seek(0, 2)
    f.write('%s\t%s\t%s\n' % (quote(team, ''),
                              quote(passwd, ''),
                              quote(color, '')))

def color(team):
    validate(team)
    t = teams.get(team)
    if not t:
        return '888888'
    return t[1]
