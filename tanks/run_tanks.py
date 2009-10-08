#! /usr/bin/python

import optparse
import shutil
import time
from tanks import Pflanzarr

T = 60*5
MAX_HIST = 30
HIST_STEP = 100 

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

    path = os.path.join(args[0], 'results')
    files = os.listdir(path)
    gameNums = []
    for file in files:
        try:
            gameNums.append( int(file) )
        except:
            continue

    gameNums.sort(reverse=True)
    highest = gameNums[0]
    for num in gameNums:
        if highest - MAX_HIST > num and not (num % HIST_STEP == 0):
            shutil.rmtree(os.path.join(path, num))

    diff = time.time() - start
    if diff - T > 0:
        time.sleep( diff - T )
