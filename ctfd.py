#! /usr/bin/env python3

import asyncore
import pointsd
import game
import flagd
import histogram
import config
import os
import sys

do_reap = False

def chart(s):
    if not os.fork():
        histogram.main(s)
        sys.exit(0)

def reap():
    try:
        while True:
            os.waitpid(0, os.WNOHANG)
    except OSError:
        pass

def sigchld(signum, frame):
    do_reap = True

def main():
    pointsrv = pointsd.start()
    flagsrv = flagd.start()

    s = pointsrv.store
    slen = 0
    while True:
        if do_reap:
            reap()
        asyncore.loop(timeout=30, use_poll=True, count=1)
        if len(s) > slen:
            slen = len(s)
            chart(s)

main()
