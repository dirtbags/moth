import asynchat
import asyncore
import socket

class Flagger(asynchat.async_chat):
    """Connection to flagd"""

    def __init__(self, addr, auth):
        asynchat.async_chat.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect((addr, 6668))
        self.push(auth + b'\n')
        self.flag = None

    def handle_read(self):
        msg = self.recv(4096)
        raise ValueError("Flagger died: %r" % msg)

    def handle_error(self):
        # If we lose the connection to flagd, nobody can score any
        # points.  Terminate everything.
        asyncore.close_all()
        asynchat.async_chat.handle_error(self)

    def set_flag(self, team):
        if team:
            eteam = team.encode('utf-8')
        else:
            eteam = b''
        self.push(eteam + b'\n')
        self.flag = team
