#! /usr/bin/python

import asynchat
import asyncore
import socket

class Flagger(asynchat.async_chat):
    """Use to connect to flagd and submit the current flag holder."""

    def __init__(self, addr, auth):
        asynchat.async_chat.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect((addr, 1))
        self.push(auth + '\n')
        self.flag = None

    def handle_read(self):
        # We don't care.
        msg = self.recv(4096)

    def handle_error(self):
        # If we lose the connection to flagd, nobody can score any
        # points.  Terminate everything.
        asyncore.close_all()
        asynchat.async_chat.handle_error(self)

    def set_flag(self, team):
        if team:
            eteam = team.encode('utf-8')
        else:
            eteam = ''
        self.push(eteam + '\n')
        self.flag = team
