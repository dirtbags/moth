#! /usr/bin/env python3

import asyncore
import os
import sys
import optparse
import signal
from ctf import pointsd
from ctf import flagd
from ctf import histogram
from ctf import config

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
    p = optparse.OptionParser()
    p.add_option('-p', '--genpass', dest='cat', default=None,
                 help='Generate a flagger password for the given category')
    opts, args = p.parse_args()
    if opts.cat:
        print('%s:::%s' % (opts.cat, flagd.hexdigest(opts.cat.encode('utf-8'))))
        return

    pointsrv = pointsd.start()
    flagsrv = flagd.start()

    signal.signal(signal.SIGCHLD, sigchld)
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
