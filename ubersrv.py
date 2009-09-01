#! /usr/bin/env python3

import asyncore
import pointsd
import flagd

def main():
    pointsd.start()
    flagd.start()
    asyncore.loop(timeout=30, use_poll=True)

main()
