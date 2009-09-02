#! /usr/bin/env python3

import socket
import hmac
import struct
import io
import teams

##
## Authentication
##

key = b'mollusks peck my galloping genitals'
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
def encode_request(id, when, cat, team, score):
    base = (struct.pack('!II', id, when) +
            packstr(cat) +
            packstr(team) +
            struct.pack('!i', score))
    return sign(base)

def decode_request(b):
    base = check_sig(b)
    id, when, base = unpack('!II', base)
    cat, base = unpackstr(base)
    team, base = unpackstr(base)
    score, base = unpack('!i', base)
    assert not base
    return (id, when, cat, team, score)


##
## Response
##
def encode_response(id, txt):
    base = (struct.pack('!I', id) +
            packstr(txt))
    return sign(base)

def decode_response(b):
    base = check_sig(b)
    id, base = unpack('!I', base)
    txt, base = unpackstr(base)
    assert not base
    return (id, txt)


##
## Storage
##
def incdict(dict, key, amt=1):
    dict[key] = dict.get(key, 0) + amt

class Storage:
    def __init__(self, fn):
        self.points_by_team = {}
        self.points_by_cat = {}
        self.points_by_cat_team = {}
        self.log = []
        self.f = io.BytesIO()

        # Read stored scores
        try:
            f = open(fn, 'rb')
            while True:
                l = f.read(4)
                if not l:
                    break
                (l,) = struct.unpack('!I', l)
                b = f.read(l)
                when, score, catlen, teamlen, b = unpack('!IiHH', b)
                cat = b[:catlen].decode('utf-8')
                team = b[catlen:].decode('utf-8')
                req = (when, cat, team, score)
                self.add(req, False)
            f.close()
        except IOError:
            pass

        try:
            self.f = open(fn, 'ab')
        except IOError:
            self.f = None

    def __len__(self):
        return len(self.log)

    def add(self, req, write=True):
        when, cat, team, score = req

        incdict(self.points_by_team, team, score)
        incdict(self.points_by_cat, cat, score)
        incdict(self.points_by_cat_team, (cat, team), score)
        self.log.append(req)

        if write:
            cat = cat.encode('utf-8')
            team = team.encode('utf-8')
            b = (struct.pack('!IiHH', when, score, len(cat), len(team)) +
                 cat + team)
            lb = struct.pack('!I', len(b))
            self.f.write(lb)
            self.f.write(b)
            self.f.flush()

    def categories(self):
        return sorted(self.points_by_cat)

    def teams(self):
        return sorted(self.points_by_team)

    def cat_points(self, cat):
        return self.points_by_cat.get(cat, 0)

    def team_points(self, team):
        return self.points_by_team.get(team, 0)

    def team_points_in_cat(self, cat, team):
        return self.points_by_cat_team.get((cat, team), 0)


##
## Colors
##
def colors(teams):
    colors = ['F0888A', '88BDF0', '00782B', '999900', 'EF9C00',
              'F4B5B7', 'E2EFFB', '89CA9D', 'FAF519', 'FFE7BB',
              'BA88F0', '8DCFF4', 'BEDFC4', 'FFFAB2', 'D7D7D7',
              'C5B9D7', '006189', '8DCB41', 'FFCC00', '898989']
    return dict(zip(teams, colors))




##
## Testing
##

def test():
    import time
    import os

    now = int(time.time())

    req = (now, 'category 5', 'foobers in heat', 43)
    assert decode_request(encode_request(*req)) == req

    rsp = (now, 'cat6', 'hello world')
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
    assert s.team_points_in_cat('cat1', 'aardvarks') == 10
    assert s.team_points_in_cat('merf', 'zebras') == 0

    del s
    s = Storage('test.dat')
    assert s.teams() == ['aardvarks', 'zebras']

    print('all tests pass; output file is test.dat')


if __name__ == '__main__':
    test()


