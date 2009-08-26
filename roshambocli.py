#! /usr/bin/env python3

import socket
import json

class Client:
    rbufsize = -1
    wbufsize = 0

    def __init__(self, addr):
        self.conn = socket.create_connection(addr)
        self.wfile = self.conn.makefile('wb', self.wbufsize)
        self.rfile = self.conn.makefile('rb', self.rbufsize)

    def write(self, *val):
        s = json.dumps(val)
        print('--> %s' % s)
        self.wfile.write(s.encode('utf-8') + b'\n')

    def read(self):
        line = self.rfile.readline().strip().decode('utf-8')
        if not line:
            return
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
            print(ret)

def main():
    c = Client(('localhost', 5388))
    c.command('^', 'lobby')
    c.command('login', 'zebra', 'furble')
    c.command('rock')

main()
