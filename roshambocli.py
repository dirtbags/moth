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
            print('--> %s' % s)
        self.wfile.write(s.encode('utf-8') + b'\n')

    def read(self):
        line = self.rfile.readline().strip().decode('utf-8')
        if not line:
            return
        if self.debug:
            print ('<-- %s' % line)
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


class RandomBot(threading.Thread):
    def __init__(self, team):
        threading.Thread.__init__(self)
        self.team = team

    def run(self):
        c = Client(('localhost', 5388))
        #print('lobby', c.command('^', 'lobby'))
        c.command('login', self.team, 'furble')
        while True:
            move = random.choice(['rock', 'scissors', 'paper'])
            ret = c.command(move)
            if ret == ['WIN']:
                print('%s wins' % self.team)
            time.sleep(random.uniform(0.2, 2))

def main():
    bots = []
    for i in ['zebra', 'aardvark', 'wembly']:
        bots.append(RandomBot(i).start())

main()
