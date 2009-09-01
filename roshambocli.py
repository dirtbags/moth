#! /usr/bin/env python3

import socket
import json
import random
import time
import threading

class Client:
    rbufsize = -1
    wbufsize = 0
    debug = False

    def __init__(self, addr):
        self.conn = socket.create_connection(addr)
        self.wfile = self.conn.makefile('wb', self.wbufsize)
        self.rfile = self.conn.makefile('rb', self.rbufsize)

    def write(self, *val):
        s = json.dumps(val)
        if self.debug:
            print(self, '--> %s' % s)
        self.wfile.write(s.encode('utf-8') + b'\n')

    def read(self):
        line = self.rfile.readline().strip().decode('utf-8')
        if not line:
            return
        if self.debug:
            print (self, '<-- %s' % line)
        return json.loads(line)

    def command(self, *val):
        self.write(*val)
        ret = self.read()
        if ret[0] == 'OK':
            return ret[1]
        elif ret[0] == 'ERR':
            raise ValueError(ret[1])
        else:
            return ret


class IdiotBot(threading.Thread):
    def __init__(self, team, move):
        threading.Thread.__init__(self)
        self.team = team
        self.move = move

    def get_move(self):
        return self.move

    def run(self):
        c = Client(('localhost', 5388))
        c.debug = False
        #print('lobby', c.command('^', 'lobby'))
        c.command('login', self.team, 'furble')
        while True:
            move = self.get_move()
            ret = c.command(move)
            if ret == ['WIN']:
                print('%s wins' % self.team)
            amt = random.uniform(0.1, 1.2)
            if c.debug:
                print(c, 'sleep %f' % amt)
            time.sleep(amt)

def main():
    bots = []
    for team, move in (('rockbot', 'rock'), ('cutbot', 'scissors'), ('paperbot', 'paper')):
        bots.append(IdiotBot(team, move).start())

main()
