#!/usr/bin/python3.0

from ctf.flagd import Flagger

key = 'tanks:::2bac5e912ff2e1ad559b177eb5aeecca'

f = Flagger.Flagger('localhost', key)

while 1:
    time.sleep(1)
    try:
        winner = open('/var/lib/tanks/winner').read()
    except:
        winner = None

    f.set_flag(winner)
