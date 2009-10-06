import math

def rotatePoint(point, angle):
    """Assuming 0,0 is the center, rotate the given point around it."""

    x,y = point
    r = math.sqrt(x**2 + y**2)
    if r == 0:
        return 0, 0

    theta = math.acos(x/r)
    if y < 0:
        theta = -theta
    theta = theta + angle
    return int(round(r*math.cos(theta))), int(round(r*math.sin(theta)))

def rotatePoly(points, angle):
    """Rotate the given list of points around 0,0 by angle."""
    return [ rotatePoint(point, angle) for point in points ]

def displace(point, disp, limits):
    """Displace point by disp, wrapping around limits."""
    x = (point[0] + disp[0])
    while x >= limits[0]:
        x = x - limits[0]
    while x < 0:
        x = x + limits[0]

    y = (point[1] + disp[1])
    while y >= limits[1]:
        y = y - limits[1]
    while y < 0:
        y = y + limits[1]

    return x,y

def displacePoly(points, disp, limits, coordSequence=False):
    """Displace each point (x,y) in 'points' by 'disp' (x,y).  The limits of 
    the drawing space are assumed to be at x=0, y=0 and x=limits[0], 
    y=limits[1].  If the poly overlaps the edge of the drawing space, the
    poly is duplicated on each side.
@param coordSequence: If true, the coordinates are returned as a sequence -
                      x1, y1, x2, y2, ...  This is need by some PIL drawing 
                      commands.
@returns: A list of polys, displaced by disp
    """
    xDup = 0; yDup = 0
    maxX, maxY = limits
    basePoints = []
    for point in points:
        x,y = point[0] + disp[0], point[1] + disp[1]

        # Check if duplication is needed on each axis
        if x > maxX:
            # If this is negative, then we need to duplicate in the negative
            # direction.
            xDup = -1
        elif x < 0:
            xDup = 1

        if y > maxY:
            yDup = -1
        elif y < 0:
            yDup = 1

        basePoints.append( (x,y) )

    polys = [basePoints]
    if xDup:
        polys.append([(x + maxX*xDup, y) for x,y in basePoints] )
    if yDup:
        polys.append([(x, maxY*yDup + y) for x,y in basePoints] )
    if xDup and yDup:
        polys.append([(x+maxX*xDup, maxY*yDup+y) for x,y in basePoints])

    # Switch coordinates to sequence mode.
    # (x1, y1, x2, y2) instead of ((x1, y1), (x2, y2))
    if coordSequence:
        seqPolys = []
        for poly in polys:
            points = []
            for point in poly:
                points.extend(point)
            seqPolys.append(points)
        polys = seqPolys

    return polys

def polar2cart(r, theta):
    """Return the cartesian coordinates for r, theta."""
    x = r*math.cos(theta)
    y = r*math.sin(theta)
    return x,y

def minShift(center, point, limits):
    """Get the minimum distances between the two points, given that the board
    wraps at the givin limits."""
    dx = point[0] - center[0]
    if dx < -limits[0]/2.0:
        dx = point[0] + limits[0] - center[0]
    elif dx > limits[0]/2.0:
        dx = point[0] - (center[0] + limits[0])

    dy = point[1] - center[1]
    if dy < - limits[1]/2.0:
        dy = point[1] + limits[1] - center[1]
    elif dy > limits[1]/2.0:
        dy = point[1] - (limits[1] + center[1])

    return dx, dy

def relativePolar(center, point, limits):
    """Returns the angle, from zero, to the given point assuming this
center is the origin. Take into account wrapping round the limits of the board.
@returns: r, theta
    """

    dx, dy = minShift(center, point, limits)

    r = math.sqrt(dx**2 + dy**2)
    theta = math.acos(dx/r)
    if dy < 0:
        theta = 2*math.pi - theta

    return r, theta

def reduceAngle(angle):
    """Reduce the angle such that it is in 0 <= angle < 2pi"""

    while angle >= math.pi*2:
        angle = angle - math.pi*2
    while angle < 0:
        angle = angle + math.pi*2

    return angle

def angleDiff(angle1, angle2):
    """Returns the difference between the two angles.  They are assumed 
to be in radians, and must be in the range 0 <= angle < 2*pi.
@raises AssertionError: The angles given must be in the range 0 <= angle < 2pi
@returns: The minimum distance between the two angles;  The distance
          is negative if angle2 leads angle1 (clockwise)..
    """

    for angle in angle1, angle2:
        assert angle < 2*math.pi and angle >= 0, \
               'angleDiff: bad angle %s' % angle

    diff = angle2 - angle1
    if diff > math.pi:
        diff = diff - 2*math.pi
    elif diff < -math.pi:
        diff = diff + 2*math.pi

    return diff

def getDist(point1, point2):
    """Returns the distance between point1 and point2."""
    dx = point2[0] - point1[0]
    dy = point2[1] - point1[1]

    return math.sqrt(dx**2 + dy**2)

def segmentCircleCollision(segment, center, radius):
    """Return True if the given circle touches the given line segment.
@param segment: A list of two points [(x1,y1), (x2, y2)] that define
                the line segment.
@param center: The center point of the circle.
@param radius: The radius of the circle.
@returns: True if the the circle touches the line segment, False otherwise.
    """

    a = getDist(segment[0], center)
    c = getDist(segment[1], center)
    base = getDist(segment[0], segment[1])

    # If we're close enough to the end points, then we're close
    # enough to the segment.
    if a < radius or c < radius:
        return True

    # First we find the are of the triangle formed by the line segment
    # and point. I use Heron's formula for the area.  Using this, we'll
    # find the distance d from the point to the line.  We'll later make
    # sure that the collision is with the line segment, and not just the 
    # line.
    s = (a + c + base)/2
    A = math.sqrt(s*(s - a)*(s - c)*(s - base))
    d = 2*A/base

#        print s, a, c, A, d, radius

    # If the distance from the point to the line is more than the
    # target radius, this isn't a hit.
    if d > radius:
        return False

    # If the distance from an endpoint to the intersection between 
    # our line segment and the line perpendicular to it that passes through
    # the point is longer than the line segment, then this isn't a hit.
    elif math.sqrt(a**2 - d**2) > base or \
         math.sqrt(c**2 - d**2) > base:
        return False
    else:
        # The triangle is acute, that means we're close enough.
        return True
