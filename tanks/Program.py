#! /usr/bin/python

import forf
import random
import rfc822
from cStringIO import StringIO
from math import pi

def deg2rad(deg):
    return float(deg) * pi / 180

def rad2deg(rad):
    return int(rad * 180 / pi)

class Environment(forf.Environment):
    def __init__(self, tank, stdout):
        forf.Environment.__init__(self)
        self.tank = tank
        self.stdout = stdout

    def err(self, msg):
        self.stdout.write('Error: %s\n' % msg)

    def msg(self, msg):
        self.stdout.write('%s\n' % msg)

    def cmd_random(self, data):
        high = data.pop()
        ret = random.randrange(high)
        data.push(ret)

    def cmd_fireready(self, data):
        ret = self.tank.fireReady()
        data.push(ret)

    def cmd_sensoractive(self, data):
        sensor = data.pop()
        try:
            ret = int(self.tank.getSensorState(sensor))
        except KeyError:
            ret = 0
        data.push(ret)

    def cmd_getturret(self, data):
        rad = self.tank.getTurretAngle()
        deg = rad2deg(rad)
        data.push(deg)

    def cmd_setled(self, data):
        self.tank.setLED()

    def cmd_fire(self, data):
        self.tank.setFire()

    def cmd_move(self, data):
        right = data.pop()
        left = data.pop()
        self.tank.setMove(left, right)

    def cmd_setturret(self, data):
        deg = data.pop()
        rad = deg2rad(deg)
        self.tank.setTurretAngle(rad)


class Program:
    def __init__(self, tank, source):
        self.tank = tank
        self.stdout = StringIO()
        self.env = Environment(self.tank, self.stdout)

        code_str = self.read_source(StringIO(source))
        self.env.parse_str(code_str)

    def get_output(self):
        return self.stdout.getvalue()

    def read_source(self, f):
        """Read in a tank program, establish sensors, and return code.

        Tank programs are stored as rfc822 messages.  The header
        block includes fields for sensors (Sensor:)
        and other crap which may be used later.
        """

        message = rfc822.Message(f)
        print 'reading tank %s' % message['Name']
        sensors = message.getallmatchingheaders('Sensor')
        for s in sensors:
            k, v = s.strip().split(':')
            r, angle, width, turret = [int(p) for p in v.split()]
            r = float(r) / 100
            angle = deg2rad(angle)
            width = deg2rad(width)
            self.tank.addSensor(r, angle, width, turret)
        return message.fp.read()

    def run(self):
        self.env.eval()

