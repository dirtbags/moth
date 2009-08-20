#! /usr/bin/env python3

import socketserver
import threading
import queue
import time
import hmac
import optparse
import points
import pointscli

key = b'My First Shared Secret (tm)'
def hexdigest(data):
    return hmac.new(key, data).hexdigest()

house = 'dirtbags'              # House team name
flags = {}
toscore = queue.Queue(50)

class Submitter(threading.Thread):
    def run(self):
        self.sock = pointscli.makesock('localhost')
        while True:
            try:
                delay = 60 - (time.time() % 60)
                cat, team = toscore.get(True, delay)
                self.submit(cat, team)
            except queue.Empty:
                self.once()

    def once(self):
        global flags
        global toscore

        for cat, team in flags.items():
            self.submit(cat, team)

    def submit(self, cat, team):
        try:
            pointscli.submit(self.sock, cat, team, 1)
        except:
            print('Uh oh, exception submitting')


class CategoryHandler(socketserver.StreamRequestHandler):
    def handle(self):
        global flags

        try:
            catpass = self.rfile.readline().strip()
            cat, passwd = catpass.split(b':::')
            passwd = passwd.decode('utf-8')
            if passwd != hexdigest(cat):
                self.wfile.write(b'ERROR :Closing Link: Invalid password\n')
                return
            cat = cat.decode('utf-8')
        except ValueError as foo:
            self.wfile.write(b'ERROR :Closing Link: Invalid command\n')
            return

        flags[cat] = house
        while True:
            team = self.rfile.readline().strip().decode('utf-8')
            if not team:
                break
            flags[cat] = team
            toscore.put((cat, team)) # score a point immediately
        flags[cat] = house

class MyServer(socketserver.ThreadingTCPServer):
    allow_reuse_address = True


def main():
    p = optparse.OptionParser()
    p.add_option('-p', '--genpass', dest='cat', default=None,
                 help='Generate a password for the given category')
    opts, args = p.parse_args()
    if opts.cat:
        print('%s:::%s' % (opts.cat, hexdigest(opts.cat.encode('utf-8'))))
        return

    submitter = Submitter()
    submitter.start()
    server = MyServer(('', 6668), CategoryHandler)
    server.serve_forever()


if __name__ == '__main__':
    main()
