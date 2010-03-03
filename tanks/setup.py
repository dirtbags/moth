"""Each of these classes provides a function for configuring a tank.  
They should inherit from Function.Function.  
To make one available to the tank programmer, add it to the dictionary at
the end of this file."""

import Function

class AddSensor(Function.Function):
    """addsensor(range, angle, width, [turretAttached])
Add a new sensor to the tank.  Sensors are an arc (pie slice) centered on 
the tank that detect other tanks within their sweep.  
A sensor is 'on' if another tank is within this arc.
Sensors are numbered, starting at 0, in the order they are added.
<p>
range - The range of the sensor, as a percent of the tanks max range.
angle - The angle of the center of the sensor, in degrees.
width - The width of the sensor, in percent (100 is a full circle).
turretAttached - Normally, the angle is relative to the front of the
tank.  When this is set, the angle is relative to the current turret 
direction.
<p>
Sensors are drawn for each tank, but not in the way you might expect.
Instead of drawing a pie slice (the actual shap of the sensor), an arc with 
the end points connected by a line is drawn. Sensors with 0 width don't show
up, but still work.
"""

    def __init__(self, range, angle, width, turretAttached=False):
        
        self._checkRange(range, 'sensor range')

        self._range = range / 100.0
        self._width = self._convertAngle(width, 'sensor width')
        self._angle = self._convertAngle(angle, 'sensor angle')
        self._turretAttached = turretAttached

    def __call__(self, tank):
        tank.addSensor(self._range, self._angle, self._width, 
                       self._turretAttached)

class AddToggle(Function.Function):
    """addtoggle([state])
Add a toggle to the tank.  The state of the toggle defaults to 0 (False).
These essentially act as a single bit of memory.
Use the toggle() condition to check its state and the settoggle, cleartoggle,
and toggle actions to change the state.  Toggles are named numerically, 
starting at 0.
"""
    def __init__(self, state=0):
        self._state = state

    def __call__(self, tank):
        if len(tank.toggles) >= tank.SENSOR_LIMIT:
            raise ValueError('You can not have more than 10 toggles.')

        tank.toggles.append(self._state)

class AddTimer(Function.Function):
    """addtimer(timeout)
Add a new timer (they're numbered in the order added, starting from 0), 
with the given timeout.  The timeout is in number of turns.  The timer
is created in inactive mode.  You'll need to do a starttimer() action
to reset and start the timer.  When the timer expires, the timer()
condition will begin to return True."""
    def __init__(self, timeout):
        self._timeout = timeout
    def __call__(self, tank):
        tank.addTimer(timeout)

setup = {'addsensor': AddSensor,
         'addtoggle': AddToggle,
         'addtimer': AddTimer}
