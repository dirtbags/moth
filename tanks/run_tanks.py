#! /usr/bin/python

import asynchat
import asyncore
import optparse
import os
import shutil
import socket
import time
from tanks import Pflanzarr

T = 60*5
MAX_HIST = 30
HIST_STEP = 100
key = 'tanks:::2bac5e912ff2e1ad559b177eb5aeecca'

class Flagger(asynchat.async_chat):
    """Use to connect to flagd and submit the current flag holder."""

    def __init__(self, addr, auth):
        asynchat.async_chat.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect((addr, 6668))
        self.push(auth + '\n')
        self.flag = None

    def handle_read(self):
        msg = self.recv(4096)
        raise ValueError("Flagger died: %r" % msg)

    def handle_error(self):
        # If we lose the connection to flagd, nobody can score any
        # points.  Terminate everything.
        asyncore.close_all()
        asynchat.async_chat.handle_error(self)

    def set_flag(self, team):
        if team:
            eteam = team
        else:
            eteam = ''
        self.push(eteam + '\n')
        self.flag = team


def run_tanks(args, turns, flagger):
    p = Pflanzarr.Pflanzarr(args[0], args[1])
    p.run(turns)

    path = os.path.join(args[0], 'results')
    files = os.listdir(path)
    gameNums = []
    for file in files:
        try:
            gameNums.append( int(file) )
        except:
            continue

    gameNums.sort(reverse=True)
    highest = gameNums[0]
    for num in gameNums:
        if highest - MAX_HIST > num and not (num % HIST_STEP == 0):
            shutil.rmtree(os.path.join(path, str(num)))

    try:
        winner = open('/var/lib/tanks/winner').read().strip()
    except:
        winner = None
    flagger.set_flag(winner)


def main():
    parser = optparse.OptionParser('DATA_DIR easy|medium|hard MAX_TURNS')
    opts, args = parser.parse_args()
    if (len(args) != 3) or (args[1] not in ('easy', 'medium', 'hard')):
        parser.error('Wrong number of arguments')
    try:
        turns = int(args[2])
    except:
        parser.error('Invalid number of turns')


    flagger = Flagger('localhost', key)
    lastrun = 0
    while True:
        asyncore.loop(60, count=1)
        now = time.time()
        if now - lastrun >= 60:
            run_tanks(args, turns, flagger)
            lastrun = now

if __name__ == '__main__':
    main()
