import time
from tanks import Pflanzarr
import sys

T = 60*5

try:
    while 1:
        start = time.time()
        p = Pflanzarr(sys.argv[1], sys.argv[2])
        p.run(int(sys.argv[3]))
        
        diff = time.time() - start
        if diff - T > 0:
            time.sleep( diff - T )

except:
    print 'Usage: python2.6 run_tanks.py data_dir easy|medium|hard max_turns'

    
