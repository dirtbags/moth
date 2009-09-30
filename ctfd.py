#! /usr/bin/env python3

import asyncore
import pointsd
import game
import flagd
import histogram
import config

def main():
    pointsrv = pointsd.start()
    flagsrv = flagd.start()

    if config.enabled('roshambo'):
        import roshambo
        roshambosrv = roshambo.start()

    s = pointsrv.store
    slen = 0
    while True:
        asyncore.loop(timeout=30, use_poll=True, count=1)
        if len(s) > slen:
            slen = len(s)
            histogram.main(s)

main()
