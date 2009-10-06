"""Define new condition functions here.  Add it to the conditions dictionary
at the end to make it usable by Program.Program.  These should inherit from 
Function.Function."""

import Function
import random

class Sense(Function.Function):
    """sense(#, [invert])
    Takes a Sensor number as an argument.  
    Returns True if the given sensor is currently activated, False otherwise.
    If the option argument invert is set to true then logic is inverted,
    and then sensor returns True when it is NOT activated, and False when 
    it is.  Invert is false by default.
    """
    
    def __init__(self, sensor, invert=0):
        self._sensor = sensor
        self._invert = invert

    def __call__(self, tank):
        state = tank.getSensorState(self._sensor)
        if self._invert:
            return not state
        else:
            return state

class Toggle(Function.Function):
    """toggle(#)
Returns True if the given toggle is set, False otherwise. """
    def __init__(self, toggle):
        self._toggle = toggle
    def __call__(self, tank):
        return tank.toggles[toggle]

class TimerCheck(Function.Function):
    """timer(#, [invert])
Checks the state of timer # 'key'.  Returns True if time has run out.
If invert is given (and true), then True is returned if the timer has
yet to expire.
"""
    def __init__(self, key, invert=0):
        self._key = key
        self._invert = invert
    def __call__(self, tank):
        state = tank.checkTimer(self._key)
        if invert:
            return not state
        else:
            return state

class Random(Function.Function):
    """random(n,m)
    Takes two arguments, n and m.  Generates a random number between 1
    and m (inclusive) each time it's checked.  If the random number is less 
    than or equal
    to n, then the condition returns True.  Returns False otherwise."""

    def __init__(self, n, m):
        self._n = n
        self._m = m

    def __call__(self, tank):
        if random.randint(1,self._m) <= self._n:
            return True
        else:
            return False

class Sin(Function.Function):
    """sin(T)
    A sin wave of period T (in turns).  Returns True when the wave is positive.
    A wave with period 1 or 2 is always False (it's 0 each turn), only
    at periods of 3 or more does this become useful."""

    def __init__(self, T):
        self._T = T

    def __call__(self, tank):
        turn = tank.turn
        factor = math.pi/self._T
        if math.sin(turn * factor) > 0:
            return True
        else:
            return False

class Cos(Function.Function):
    """cos(T)
    A cos wave with period T (in turns). Returns True when the wave is 
    positive.  A wave of period 1 is always True.  Period 2 is True every
    other turn, etc."""

    def __init__(self, T):
        self._T = T

    def __call__(self, tank):
        
        turn = tank.turn
        factor = math.pi/self._T
        if math.cos(turn * factor) > 0:
            return True
        else:
            return False

class FireReady(Function.Function):
    """fireready()
    True when the tank can fire."""
    def __call__(self, tank):
        return tank.fireReady()

class FireNotReady(Function.Function):
    """firenotready()
    True when the tank can not fire."""
    def __call__(self, tank):
        return not tank.fireReady()

### When adding names to this dict, make sure they are lower case and alpha
### numeric.
conditions = {'sense': Sense,
              'random': Random,
              'toggle': Toggle,
              'sin': Sin,
              'cos': Cos,
              'fireready': FireReady,
              'firenotready': FireNotReady,
              'timer': TimerCheck }
