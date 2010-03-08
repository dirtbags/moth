import math
import random
from sets import Set as set

import GameMath as gm
import Program
from cStringIO import StringIO

class Tank(object):

    # How often, in turns, that we can fire.
    FIRE_RATE = 20
    # How far the laser shoots from the center of the tank
    FIRE_RANGE = 45.0
    # The radius of the tank, from the center of the turret.
    # This is what is used for collision and hit detection.
    RADIUS = 7.5
    # Max speed, in pixels
    MAXSPEED = 7.0
    # Max acceleration, as a fraction of speed.
    MAXACCEL = 35
    # Sensor range, in pixels
    SENSOR_RANGE = 90.0
    # Max turret turn rate, in radians
    TURRET_TURN_RATE = math.pi/10

    # The max number of sensors/timers/toggles
    SENSOR_LIMIT = 10

    def __init__(self, name, pos, color, boardSize, angle=None, tAngle=None):
        """Create a new tank.
@param name: The name name of the tank.  Stored in self.name.
@param pos: The starting position of the tank (x,y)
@param color: The color of the tank.
@param boardSize: The size of the board. (maxX, maxY)
@param angle: The starting angle of the tank, defaults to random.
@param tAngle: The starting turretAngle of the tank, defaults to random.
        """

        self.name = name

        assert len(pos) == 2 and pos[0] > 0 and pos[1] > 0, \
               'Bad starting position: %s' % str(pos)
        self.pos = pos

        # The last speed of each tread (left, right)
        self._lastSpeed = 0.0, 0.0
        # The next speed that the tank should try to attain.
        self._nextMove = 0,0

        # When set, the led is drawn on the tank.
        self.led = False

        assert len(boardSize) == 2 and boardSize[0] > 0 and boardSize[1] > 0
        # The limits of the playfield (maxX, maxY)
        self._limits = boardSize

        # The current angle of the tank.
        if angle is None:
            self._angle = random.random()*2*math.pi
        else:
            self._angle = angle

        # The current angle of the turret
        if tAngle is None:
            self._tAngle = random.random()*2*math.pi
        else:
            self._tAngle = tAngle

        self.color = color

        # You can't fire until fireReady is 0.
        self._fireReady = self.FIRE_RATE
        # Means the tank will fire at it's next opportunity.
        self._fireNow = False
        # True when the tank has fired this turn (for drawing purposes)
        self._fired = False

        # What the desired turret angle should be (from the front of the tank).
        # None means the turret should stay stationary.
        self._tGoal = None

        # Holds the properties of each sensor
        self._sensors = []
        # Holds the state of each sensor
        self._sensorState = []

        # The tanks toggle memory
        self.toggles = []

        # The tanks timers
        self._timers = []

        # Is this tank dead?
        self.isDead = False

        # Death reason
        self.deathReason = 'survived'

        # Something to log to
        self.stdout = StringIO()


    def __repr__(self):
        return '<tank: %s, (%d, %d)>' % (self.name, self.pos[0], self.pos[1])

    def fire(self, near):
        """Shoots, if ordered to and able.  Returns the set of tanks
    destroyed."""

        killed = set()
        if self._fireReady > 0:
            # Ignore the shoot order
            self._fireNow = False

        if self._fireNow:
            self._fireNow = False
            self._fireReady = self.FIRE_RATE
            self._fired = True


            firePoint = gm.polar2cart(self.FIRE_RANGE,
                                         self._angle + self._tAngle)
            for tank in near:
                enemyPos = gm.minShift(self.pos, tank.pos, self._limits)
                if gm.segmentCircleCollision(((0,0), firePoint), enemyPos,
                                             self.RADIUS):
                    killed.add(tank)
        else:
            self._fired = False

        return killed

    def addSensor(self, range, angle, width, attachedTurret=False):
        """Add a sensor to this tank.
@param angle: The angle, from the tanks front and going clockwise,
              of the center of the sensor. (radians)
@param width: The width of the sensor (radians).
@param range: The range of the sensor (percent)
@param attachedTurret: If set, the sensor moves with the turret.
        """
        assert range >=0 and range <= 1, "Invalid range value."

        if len(self._sensors) >= self.SENSOR_LIMIT:
            raise ValueError("You can only have 10 sensors.")

        range = range * self.SENSOR_RANGE

        if attachedTurret:
            attachedTurret = True
        else:
            attachedTurret = False

        self._sensors.append((range, angle, width, attachedTurret))
        self._sensorState.append(False)

    def getSensorState(self, key):
        return self._sensorState[key]

    def setMove(self, left, right):
        """Parse the speed values given, and set them for the next move."""

        self._nextMove = left, right

    def getTurretAngle(self):
        return self._tAngle

    def setTurretAngle(self, angle=None):
        """Set the desired angle of the turret. No angle means the turret
    should remain stationary."""

        if angle is None:
            self._tGoal = None
        else:
            self._tGoal = gm.reduceAngle(angle)

    def setFire(self):
        """Set the tank to fire, next turn."""
        self._fireNow = True

    def fireReady(self):
        """Returns True if the tank can fire now."""
        return self._fireReady == 0

    def setLED(self):
        self.led = True

    def set_program(self, text):
        """Set the program for this tank."""

        self.program = Program.Program(self, text)

    def execute(self):
        """Execute this tanks program."""

        self.led = False

        self.program.run()

        self._move(self._nextMove[0], self._nextMove[1])
        self._moveTurret()
        if self._fireReady > 0:
            self._fireReady = self._fireReady - 1

    def sense(self, near):
        """Detect collisions and trigger sensors.  Returns the set of
    tanks collided with, plus this one. We do both these steps at once
    mainly because all the data is available."""

        near = list(near)
        polar = []
        for tank in near:
            polar.append(gm.relativePolar(self.pos, tank.pos, self._limits))

        for sensorId in range(len(self._sensors)):
            # Reset the sensor
            self._sensorState[sensorId] = False

            dist, sensorAngle, width, tSens = self._sensors[sensorId]

            # Adjust the sensor angles according to the tanks angles.
            sensorAngle = sensorAngle + self._angle
            # If the angle is tied to the turret, add that too.
            if tSens:
                sensorAngle = sensorAngle + self._tAngle

            while sensorAngle >= 2*math.pi:
                sensorAngle = sensorAngle - 2*math.pi

            for i in range(len(near)):
                r, theta = polar[i]
                # Find the difference between the object angle and the sensor.
                # The max this can be is pi, so adjust for that.
                dAngle = gm.angleDiff(theta, sensorAngle)

                rCoord = gm.polar2cart(dist, sensorAngle - width/2)
                lCoord = gm.polar2cart(dist, sensorAngle + width/2)
                rightLine = ((0,0), rCoord)
                leftLine = ((0,0), lCoord)
                tankRelPos = gm.minShift(self.pos, near[i].pos, self._limits)
                if r < (dist + self.RADIUS):
                    if abs(dAngle) <= (width/2) or \
                       gm.segmentCircleCollision(rightLine, tankRelPos,
                                                 self.RADIUS) or \
                       gm.segmentCircleCollision(leftLine, tankRelPos,
                                                 self.RADIUS):

                        self._sensorState[sensorId] = True
                        break

        # Check for collisions here, since we already have all the data.
        collided = set()
        for i in range(len(near)):
            r, theta = polar[i]
            if r < (self.RADIUS + near[i].RADIUS):
                collided.add(near[i])

        # Add this tank (a collision kills both, after all
        if collided:
            collided.add(self)

        return collided

    def die(self, reason):
        self.isDead = True
        self.deathReason = reason

    def _moveTurret(self):
        if self._tGoal is None or self._tAngle == self._tGoal:
            return

        diff = gm.angleDiff(self._tGoal, self._tAngle)

        if abs(diff) < self.TURRET_TURN_RATE:
            self._tAngle = self._tGoal
        elif diff > 0:
            self._tAngle = gm.reduceAngle(self._tAngle - self.TURRET_TURN_RATE)
        else:
            self._tAngle = gm.reduceAngle(self._tAngle + self.TURRET_TURN_RATE)

    def _move(self, lSpeed, rSpeed):

        assert abs(lSpeed) <= 100, "Bad speed value: %s" % lSpeed
        assert abs(rSpeed) <= 100, "Bad speed value: %s" % rSpeed

        # Handle acceleration
        if self._lastSpeed[0] < lSpeed and \
           self._lastSpeed[0] + self.MAXACCEL < lSpeed:
            lSpeed = self._lastSpeed[0] + self.MAXACCEL
        elif self._lastSpeed[0] > lSpeed and \
           self._lastSpeed[0] - self.MAXACCEL > lSpeed:
            lSpeed = self._lastSpeed[0] - self.MAXACCEL

        if self._lastSpeed[1] < rSpeed and \
           self._lastSpeed[1] + self.MAXACCEL < rSpeed:
            rSpeed = self._lastSpeed[1] + self.MAXACCEL
        elif self._lastSpeed[1] > rSpeed and \
           self._lastSpeed[1] - self.MAXACCEL > rSpeed:
            rSpeed = self._lastSpeed[1] - self.MAXACCEL

        self._lastSpeed = lSpeed, rSpeed

        # The simple case
        if lSpeed == rSpeed:
            fSpeed = self.MAXSPEED*lSpeed/100
            x = fSpeed*math.cos(self._angle)
            y = fSpeed*math.sin(self._angle)
            # Adjust our position
            self._reposition((x,y), 0)
            return

        # The works as follows:
        # The tank drives around in a circle of radius r, which is some
        # offset on a line perpendicular to the tank.  The distance it travels
        # around the circle varies with the speed of each tread, and is
        # such that each side of the tank moves an equal angle around the
        # circle.
        L = self.MAXSPEED * lSpeed/100.0
        R = self.MAXSPEED * rSpeed/100.0
        friction = .75 * abs(L-R)/(2.0*self.MAXSPEED)
        L = L * (1 - friction)
        R = R * (1 - friction)

        # Si is the speed of the tread on the inside of the turn,
        # So is the speed on the outside of the turn.
        # dir is to note the direction of rotation.
        if abs(L) > abs(R):
            Si = R; So = L
            dir = 1
        else:
            Si = L; So = R
            dir = -1

        # The width of the tank...
        w = self.RADIUS * 2

        # This is the angle that will determine the circle the tank travels
        # around.
#        theta = math.atan((So - Sl)/w)
        # This is the distance from the outer tread to the center of the
        # circle formed by it's movement.
        r = w*So/(So - Si)

        # The fraction of the circle traveled is equal to the speed of
        # the outer tread over the circumference of the circle.
        # Ft = So/(2*pi*r)
        # The angle traveled is equal to the Fraction traveled * 2 * pi
        # This reduces to a simple: So/r
        # We multiply it by dir to adjust for the direction of rotation
        theta = So/r * dir

        # These are the offsets from the center of the circle, given that
        # the tank is turned in some direction.  The tank is facing
        # perpendicular to the circle
        # So far everything has been relative to the outer tread.  At this
        # point, however, we need to move relative to the center of the
        # tank.  Hence the adjustment in r.
        x = -math.cos( self._angle + math.pi/2*dir ) * (r - w/2.0)
        y = -math.sin( self._angle + math.pi/2*dir ) * (r - w/2.0)

        # Now we just rotate the tank's position around the center of the
        # circle to get the change in coordinates.
        mx, my = gm.rotatePoint((x,y), theta)
        mx = mx - x
        my = my - y

        # Finally, we shift the tank relative to the playing field, and
        # rotate it by theta.
        self._reposition((mx, my), theta)

    def _reposition(self, move, angleChange):
        """Move the tank by x,y = move, and change it's angle by angle.
    I assume the tanks move slower than the boardSize."""

        x = self.pos[0] + move[0]
        y = self.pos[1] + move[1]
        self._angle = self._angle + angleChange

        if x < 0:
            x = self._limits[0] + x
        elif x > self._limits[0]:
            x = x - self._limits[0]

        if y < 0:
            y = self._limits[1] + y
        elif y > self._limits[1]:
            y = y - self._limits[1]

        self.pos = round(x), round(y)

        while self._angle < 0:
            self._angle = self._angle + math.pi * 2

        while self._angle > math.pi * 2:
            self._angle = self._angle - math.pi * 2

    def describe(self, f):
        """Output a description of this tank"""

        f.write('  ["%s",[' % self.color)
        for i in range(len(self._sensors)):
            dist, sensorAngle, width, tSens = self._sensors[i]

            f.write('[%d,%.2f,%.2f,%d],' % (dist, sensorAngle, width, tSens))
        f.write(']],\n')

    def draw(self, f):
        """Output this tank's state as JSON.

        [color, x, y, angle, turret_angle, led, fired]

        """

        if self.isDead:
            f.write(' 0,\n')
        else:
            flags = (self._fired << 0) | (self.led << 1)
            sensors = 0
            for i in range(len(self._sensorState)):
                sensors |= self._sensorState[i] << i
            f.write(' [%d,%d,%.2f,%.2f,%d,%d],\n' % (self.pos[0],
                                                     self.pos[1],
                                                     self._angle,
                                                     self._tAngle,
                                                     flags,
                                                     sensors))
