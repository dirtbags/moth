"""Define new action Functions here.  They should inherit from the 
Function.Function class.  To make an action usable, add it to the 
actions dictionary at the end of this file."""

import Function

class Move(Function.Function):
    """move(left tread speed, right tread speed)
    Set the speeds for the tanks right and left treads.  The speeds should
    be a number (percent power) between -100 and 100."""

    def __init__(self, left, right):
        self._checkRange(left, 'left tread speed', min=-100)
        self._checkRange(right, 'right tread speed', min=-100)

        self._left = left
        self._right = right

    def __call__(self, tank):
        tank.setMove(self._left, self._right)

class TurretCounterClockwise(Function.Function):
    """turretccw([percent speed])
    Rotate the turret counter clockwise as  a percentage of the max speed."""
    def __init__(self, speed=100):
        self._checkRange(speed, 'turret percent speed')
        self._speed = speed/100.0
    def __call__(self, tank): 
        tank.setTurretAngle(tank._tAngle - tank.TURRET_TURN_RATE*self._speed)

class TurretClockwise(Function.Function):
    """turretcw([percent speed])
    Rotate the turret clockwise at a rate preportional to speed."""

    def __init__(self, speed=100):
        self._checkRange(speed, 'turret percent speed')
        self._speed = speed/100.0
    def __call__(self, tank):
        tank.setTurretAngle(tank._tAngle + tank.TURRET_TURN_RATE*self._speed)

class TurretSet(Function.Function):
    """turretset([angle])
    Set the turret to the given angle, in degrees, relative to the front of
    the tank.  Angles increase counterclockwise.  
    The angle can be left out; in that case, this locks the turret 
    to it's current position."""

    def __init__(self, angle=None):
        # Convert the angle to radians
        if angle is not None:
            angle = self._convertAngle(angle, 'turret angle')
        
        self._angle = angle

    def __call__(self, tank):
        tank.setTurretAngle(self._angle)

class Fire(Function.Function):
    """fire()
    Attempt to fire the tanks laser cannon."""

    def __call__(self, tank):
        tank.setFire()

class SetToggle(Function.Function):
    """settoggle(key, state)
Set toggle 'key' to 'state'.
"""
    def __init__(self, key, state):
        self._key = key
        self._state = state
    def __call__(self, tank):
        tank.toggles[self._key] = self._state

class Toggle(Function.Function):
    """toggle(key)
Toggle the value of toggle 'key'.
"""
    def __init__(self, key):
        self._key = key
    def __call__(self, tank):
        try:
            tank.toggles[self._key] = not tank.toggles[self._key]
        except IndexError:
            raise IndexError('Invalid toggle: %d' % self._key)

class LED(Function.Function):
    """led(state)
Set the tanks LED to state (true is on, false is off). 
The led is a light that appears behind the tanks turret.  
It remains on for a single turn."""
    def __init__(self, state=1):
        self._state = state
    def __call__(self, tank):
        tank.led = self._state

class StartTimer(Function.Function):
    """starttimer(#)
Start (and reset) the given timer, but only if it is inactive.
"""
    def __init__(self, key):
        self._key = key
    def __call__(self, tank):
        tank.resetTimer(key)

class ClearTimer(Function.Function):
    """cleartimer(#)
Clear the given timer such that it is no longer active (inactive timers
are always False)."""
    def __init__(self, key):
        self._key = key
    def __call__(self, tank):
        tank.clearTimer(self._key)

### When adding names to this dict, make sure they are lower case and alpha
### numeric.
actions = {'move': Move,
           'turretccw': TurretCounterClockwise,
           'turretcw': TurretClockwise,
           'turretset': TurretSet,
           'fire': Fire,
           'settoggle': SetToggle,
           'toggle': Toggle,
           'led': LED,
           'starttimer': StartTimer,
           'cleartimer': ClearTimer}
