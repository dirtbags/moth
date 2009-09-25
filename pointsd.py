#! /usr/bin/env python3

import asyncore
import socket
import struct
import points
import time

house = 'dirtbags'

class MyHandler(asyncore.dispatcher):
    def __init__(self, port=6667):
        asyncore.dispatcher.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.bind(('', port))
        self.store = points.Storage()
        self.acked = set()
        self.outq = []

    def writable(self):
        return bool(self.outq)

    def handle_write(self):
        dgram, peer = self.outq.pop(0)
        self.socket.sendto(dgram, peer)

    def handle_read(self):
        now = int(time.time())
        dgram, peer = self.socket.recvfrom(4096)
        try:
            id, when, cat, team, score = points.decode_request(dgram)
        except ValueError as e:
            return self.respond(peer, now, str(e))
        team = team or house

        # Replays can happen legitimately.
        if not ((peer, id) in self.acked):
            if not (now - 2 < when <= now):
                return self.respond(peer, id, 'Your clock is off')
            self.store.add((when, cat, team, score))
            self.acked.add((peer, id))

        self.respond(peer, id, 'OK')

    def respond(self, peer, id, txt):
        resp = points.encode_response(id, txt)
        self.outq.append((resp, peer))


def start():
    return MyHandler()

if __name__ == "__main__":
    h = start()
    asyncore.loop()
