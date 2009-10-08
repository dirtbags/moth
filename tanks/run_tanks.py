#! /usr/bin/python

import time
import optparse
from tanks import Pflanzarr

T = 60*5

parser = optparse.OptionParser('DATA_DIR easy|medium|hard MAX_TURNS')
opts, args = parser.parse_args()
if (len(args) != 3) or (args[1] not in ('easy', 'medium', 'hard')):
    parser.error('Wrong number of arguments')
try:
    turns = int(args[2])
except:
    parser.error('Invalid number of turns')

while True:
    start = time.time()
    p = Pflanzarr.Pflanzarr(args[0], args[1])
    p.run(turns)

    diff = time.time() - start
    if diff - T > 0:
        time.sleep( diff - T )
