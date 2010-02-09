#! /usr/bin/env python3

import optparse
import select
import socket
import time
from . import points

def makesock(host, port=6667):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect((host, port))
    return s

def submit(cat, team, score, sock=None):
    if not sock:
        sock = makesock('localhost')
    begin = time.time()
    mark = int(begin)
    req = points.encode_request(1, mark, cat, team, score)
    while True:
        sock.send(req)
        r, w, x = select.select([sock], [], [], begin + 2 - time.time())
        if not r:
            break
        b = sock.recv(500)
        try:
            id, txt = points.decode_response(b)
        except ValueError:
            # Ignore invalid packets
            continue
        if id != 1:
            # Ignore wrong ID
            continue
        if txt == 'OK':
			sock.close()
            return
        else:
            raise ValueError(txt)

def main():
    p = optparse.OptionParser(usage='%prog CATEGORY TEAM SCORE')
    p.add_option('-s', '--host', dest='host', default='localhost',
                 help='Host to connect to')
    opts, args = p.parse_args()

    try:
        cat, team, score = args
        score = int(score)
    except ValueError:
        return p.print_usage()

    s = makesock(opts.host)

    try:
        submit(cat, team, score, sock=s)
    except ValueError as err:
        print(err)
        raise

if __name__ == '__main__':
    main()


