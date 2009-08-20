#! /usr/bin/env python3

import socketserver
import struct
import points
import time

acked = points.Storage('scores.dat')

class MyHandler(socketserver.BaseRequestHandler):
    def respond(self, when, txt):
        peer = self.request[1]
        resp = points.encode_response(when, txt)
        peer.sendto(resp, self.client_address)

    def handle(self):
        global acked

        now = int(time.time())
        data = self.request[0]
        peer = self.request[1]
        try:
            req = points.decode_request(data)
        except ValueError as e:
            return self.respond(now, str(e))
        when, cat, team, score = req

        # Replays can happen legitimately.
        if not req in acked:
            if not (now - 2 < when < now):
                resp = points.encode_response(when, 'Your clock is off')
                peer.sendto(resp, self.client_address)
                return

            acked.add(req)

        resp = points.encode_response(when, 'OK')
        peer.sendto(resp, self.client_address)


if __name__ == "__main__":
   server = socketserver.UDPServer(('', 6667), MyHandler)
   server.serve_forever()
