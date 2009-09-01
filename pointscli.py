#! /usr/bin/env python3

import optparse
import select
import points
import socket
import time

def submit(sock, cat, team, score):
    begin = time.time()
    mark = int(begin)
    req = points.encode_request(mark, cat, team, score)
    while True:
        sock.send(req)
        r, w, x = select.select([sock], [], [], begin + 2 - time.time())
        if not r:
            break
        b = sock.recv(500)
        try:
            when, cat_, txt = points.decode_response(b)
        except ValueError:
            # Ignore invalid packets
            continue
        if (when != mark) or (cat_ != cat):
            # Ignore wrong timestamp
            continue
        if txt == 'OK':
            return
        else:
            raise ValueError(txt)


def makesock(host):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.connect((host, 6667))
    return s

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
        submit(s, cat, team, score)
    except ValueError as err:
        print(err)
        raise

if __name__ == '__main__':
    main()


