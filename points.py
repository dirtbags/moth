#! /usr/bin/env python3

import socket
import hmac
import struct
import io


##
## Authentication
##

key = b'mullosks peck my galloping genitals'
def digest(data):
    return hmac.new(key, data).digest()

def sign(data):
    return data + digest(data)

def check_sig(data):
    base, mac = data[:-16], data[-16:]
    if mac == digest(base):
        return base
    else:
        raise ValueError('Invalid message digest')

##
## Marshalling
##

def unpack(fmt, buf):
    """Unpack buf based on fmt, return the rest as a buffer."""

    size = struct.calcsize(fmt)
    vals = struct.unpack(fmt, buf[:size])
    return vals + (buf[size:],)

def packstr(s):
    b = bytes(s, 'utf-8')
    return struct.pack('!H', len(b)) + b

def unpackstr(b):
    l, b = unpack('!H', b)
    s, b = b[:l], b[l:]
    return str(s, 'utf-8'), b


##
## Request
##
def encode_request(when, cat, team, score):
    base = (struct.pack('!I', when) +
            packstr(cat) +
            packstr(team) +
            struct.pack('!i', score))
    return sign(base)

def decode_request(b):
    base = check_sig(b)
    when, base = unpack('!I', base)
    cat, base = unpackstr(base)
    team, base = unpackstr(base)
    score, base = unpack('!i', base)
    assert not base
    return (when, cat, team, score)


##
## Response
##
def encode_response(when, txt):
    base = (struct.pack('!I', when) +
            packstr(txt))
    return sign(base)

def decode_response(b):
    base = check_sig(b)
    when, base = unpack('!I', base)
    txt, base = unpackstr(base)
    assert not base
    return (when, txt)


##
## Storage
##
class Storage:
    def __init__(self, fn):
        self.points_by_team = {}
        self.points_by_cat = {}
        self.log = []
        self.events = set()
        self.f = io.BytesIO()

        # Read stored scores
        try:
            f = open(fn, 'rb')
            while True:
                l = f.read(4)
                if not l:
                    return
                (l,) = struct.unpack('!I', l)
                b = f.read(l)
                req = decode_request(b)
                self.add(req)
            f.close()
        except IOError:
            pass

        self.f = open(fn, 'ab')

    def __contains__(self, req):
        return req in self.events

    def add(self, req):
        if req in self.events:
            return

        when, cat, team, score = req

        if team not in self.points_by_team:
            self.points_by_team[team] = 0
        self.points_by_team[team] += score
        if cat not in self.points_by_cat:
            self.points_by_cat[cat] = 0
        self.points_by_cat[cat] += score
        self.log.append(req)
        self.events.add(req)

        b = encode_request(*req)
        l = struct.pack('!I', len(b))
        self.f.write(l)
        self.f.write(b)

    def categories(self):
        return sorted(self.points_by_cat)

    def teams(self):
        return sorted(self.points_by_team)

    def cat_points(self, cat):
        return self.points_by_cat[cat]

    def team_points(self, team):
        return self.points_by_team[team]


##
## Testing
##

def test():
    import time
    import os

    now = int(time.time())

    req = (now, 'category 5', 'foobers in heat', 43)
    assert decode_request(encode_request(*req)) == req

    rsp = (now, 'hello world')
    assert decode_response(encode_response(*rsp)) == rsp


    try:
        os.unlink('test.dat')
    except OSError:
        pass

    s = Storage('test.dat')
    s.add((now, 'cat1', 'zebras', 20))
    s.add((now, 'cat1', 'aardvarks', 10))
    s.add((now, 'merf', 'aardvarks', 50))
    assert s.teams() == ['aardvarks', 'zebras']
    assert s.categories() == ['cat1', 'merf']
    assert s.team_points('aardvarks') == 60
    assert s.cat_points('cat1') == 30

    del s
    s = Storage('test.dat')
    assert s.teams() == ['aardvarks', 'zebras']

    print('all tests pass; output file is test.dat')


if __name__ == '__main__':
    test()


