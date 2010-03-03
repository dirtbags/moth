import math

class Function(object):
    """Represents a single condition or action.  This doc string is printed
    as user documentation.  You should override it to say something useful."""

    def __call__(self, tank):
        """The __call__ method should be of this basic form.  Actions
    should return None, conditions should return True or False.  Actions
    should utilize the set* methods of tanks.  Conditions can utilize the
    tanks get* methods."""
        pass

    def _limitArgs(self, args, max):
        """Raises a ValueError if there are more than max args."""
        if len(args) > max:
            raise ValueError("Too many arguments: %s" % ','.join(args))

    def _checkRange(self, value, name, min=0, max=100):
        """Check that the value is in the given range.
    Raises an exception with useful info for invalid values.  Name is used to
    let the user know which value is wrong."""
        try:
            value = int(value)
        except:
            raise ValueError("Invalid %s value: %s" % (name, value))
        assert value >= min and value <= max, "Invalid %s. %ss must be in"\
               " the %s %d-%d" % \
               (name, name.capitalize(), value, min, max)

        return value

    def _convertAngle(self, value, name):
        """Parse the given value as an angle in degrees, and return its value
    in radians. Raise useful errors.
    Name is used in the errors to describe the field."""
        try:
            angle = math.radians(value) 
        except:
            raise ValueError("Invalid %s value: %s" % (name, value))
    
        assert angle >= 0 and angle < 2*math.pi, "Invalid %s; "\
               "It be in the range 0 and 359." % name

        return angle

