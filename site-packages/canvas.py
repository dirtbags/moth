#! /usr/bin/python

import png
from array import array

class Canvas:
    def __init__(self, width, height, bg=(0,0,0)):
        self.width = width
        self.height = height

        # Build the canvas using arrays, which are way quick
        row = array('B')
        for i in xrange(self.width):
            row.extend(bg)

        self.bytes = array('B')
        for i in xrange(self.height):
            self.bytes.extend(row)

    def get(self, x, y):
        offs = ((y*self.width)+x)*3
        return self.bytes[offs:offs+3]

    def set(self, x, y, pixel):
        offs = ((y*self.width)+x)*3
        for i in range(3):
            self.bytes[offs+i] = pixel[i]

    def write(self, f):
        p = png.Writer(self.width, self.height)
        p.write_array(f, self.bytes)

if __name__ == '__main__':
    width = 800
    height = 600

    c = Canvas(width, height)
    for x in range(width):
        c.set(x, x % height, (x%256,(x*2)%256,(x*3)%256))
    c.write(open('foo.png', 'wb'))
