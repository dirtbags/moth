#! /usr/bin/env python3

import asyncore
import pointsd
import roshambo
import game
import flagd
import histogram

def main():
    pointsrv = pointsd.start()
    flagsrv = flagd.start()
    roshambosrv = roshambo.start()
    s = pointsrv.store
    slen = 0
    while True:
        asyncore.loop(timeout=30, use_poll=True, count=1)
        if len(s) > slen:
            slen = len(s)
            histogram.main(s)

main()
