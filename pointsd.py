#! /usr/bin/env python3

import asyncore
import socket
import struct
import points
import time

class MyHandler(asyncore.dispatcher):
    def __init__(self, port):
        asyncore.dispatcher.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.bind(('', port))
        self.acked = points.Storage('scores.dat')
        self.outq = []

    def writable(self):
        return bool(self.outq)

    def handle_write(self):
        dgram, peer = self.outq.pop(0)
        self.socket.sendto(dgram, peer)

    def handle_read(self):
        dgram, peer = self.socket.recvfrom(4096)
        now = int(time.time())
        try:
            req = points.decode_request(dgram)
        except ValueError as e:
            return self.respond(now, str(e))
        when, cat, team, score = req

        # Replays can happen legitimately.
        if not req in self.acked:
            if not (now - 2 < when <= now):
                resp = points.encode_response(when, 'Your clock is off')
                self.outq.append((resp, peer))
                return
            self.acked.add(req)

        resp = points.encode_response(when, 'OK')
        self.outq.append((resp, peer))

    def respond(self, peer, when, txt):
        resp = points.encode_response(when, txt)
        self.outq.append((resp, peer))


if __name__ == "__main__":
    h = MyHandler(6667)
    asyncore.loop()
