import optparse
import select
import points
import socket
import time

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

    now = int(time.time())
    req = points.encode_request(now, cat, team, score)
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    while True:
        s.sendto(req, (opts.host, 6667))
        r, w, x = select.select([s], [], [], 0.2)
        if r:
            b = s.recv(500)
            when, txt = points.decode_response(b)
            assert when == now
            if txt == 'OK':
                return
            print(txt)
            raise ValueError(txt)


if __name__ == '__main__':
    main()


