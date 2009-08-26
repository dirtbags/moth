#! /usr/bin/env python3

##
## XXX: Add timeout for Player if not blocked
## XXX: What if someone disconnects?
##

import json
import asyncore
import asynchat
import socket
import traceback

class Listener(asyncore.dispatcher):
    def __init__(self, addr, player_factory, manager):
        asyncore.dispatcher.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.set_reuse_addr()
        self.bind(addr)
        self.listen(4)
        self.player_factory = player_factory
        self.manager = manager

    def handle_accept(self):
        conn, addr = self.accept()
        player = self.player_factory(conn, self.manager)
        # We don't need to keep the player, asyncore.socket_map already
        # has a reference to it for as long as it's open.


class Flagger(asynchat.async_chat):
    """Connection to flagd"""

    def __init__(self, addr, auth):
        asynchat.async_chat.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect(addr)
        self.push(auth)
        self.flag = None

    def handle_read(self):
        msg = self.recv(4096)
        raise ValueError("Flagger died: %r" % msg)

    def handle_error(self):
        # If we lose the connection to flagd, nobody can score any
        # points.  Terminate everything.
        asynchat.async_chat.handle_error(self)
        asyncore.close_all()

    def set_flag(self, team):
        self.push(b'%s\n' % (team.encode('utf-8')))
        self.flag = team


class Manager:
    """Contest manager.

    When a player connects and registers, they enter the lobby.  As soon
    as there are enough players in the lobby to run a game, everyone in
    the lobby becomes a contestant.  Contestants are assigned to games.
    When the game declares a winner, the winner is added back to the list
    of contestants, and other players are sent back to the lobby.  When
    a winner is declared by the last running game, that winner gets the
    flag.

    """

    def __init__(self, nplayers, game_factory, flagger):
        self.nplayers = nplayers
        self.game_factory = game_factory
        self.flagger = flagger
        self.games = {}
        self.lobby = []
        self.contestants = []

    def enter_lobby(self, player):
        if not player.connected:
            return
        self.lobby.append(player)
        if (not self.contestants) and (len(self.lobby) >= self.nplayers):
            # If there are no contestants, the current contest has ended
            # and we're ready for a new one.
            self.contestants = self.lobby[:]
            self.run_contest()

    def add_contestant(self, player):
        self.contestants.append(player)
        self.run_contest()

    def run_contest(self):
        while len(self.contestants) >= self.nplayers:
            players = self.contestants[:self.nplayers]
            del self.contestants[:self.nplayers]
            game = self.game_factory(self, players)
            self.games[game] = players
            for player in players:
                player.attach_game(game)

    def declare_winner(self, game, winner):
        players = self.games[game]
        del self.games[game]
        players.remove(winner)
        for p in players:
            # Losers go back to the lobby
            p.lose()
            self.enter_lobby(p)
        if not self.games:
            # All games have ended and winner is the last player
            # standing.  They get the flag.
            print('%r has the flag.' % winner)
            winner.win(True)
            self.flagger.set_flag(winner.name)
            self.enter_lobby(winner)
        else:
            # Winner stays in the contest
            winner.win()
            self.add_contestant(winner)

    def player_cmd(self, args):
        cmd = args[0].lower()
        if cmd == 'lobby':
            return [p.name for p in self.lobby]
        elif cmd == 'games':
            return [[p.name for p in ps] for ps in self.games.values()]
        elif cmd == 'flag':
            return self.flagger.flag
        else:
            raise ValueError('Unrecognized manager command')


class Player(asynchat.async_chat):
    def __init__(self, sock, manager):
        asynchat.async_chat.__init__(self, sock=sock)
        self.manager = manager
        self.game = None
        self.set_terminator(b'\n')
        self.inbuf = []
        self.blocked = None
        self.name = None
        self.pending = None

    def readable(self):
        return (not self.blocked) and asynchat.async_chat.readable(self)

    def block(self):
        """Block reads"""
        self.blocked = True

    def unblock(self):
        """Unblock reads"""
        self.blocked = False

    def attach_game(self, game):
        self.game = game
        if self.pending:
            self.unblock()
            self.game.handle(self, *self.pending)

    def _write_val(self, val):
        s = json.dumps(val) + '\n'
        self.push(s.encode('utf-8'))

    def write(self, val):
        self._write_val(['OK', val])

    def err(self, msg):
        self._write_val(['ERR', msg])

    def win(self, flag=False):
        val = ['WIN']
        if flag:
            val.append('You have the flag')
        self._write_val(val)
        self.unblock()

    def lose(self):
        self._write_val(['LOSE'])
        self.unblock()

    def collect_incoming_data(self, data):
        self.inbuf.append(data)

    def found_terminator(self):
        try:
            data = b''.join(self.inbuf)
            self.inbuf = []
            val = json.loads(data.decode('utf-8'))
            cmd, args = val[0].lower(), val[1:]

            if cmd == 'login':
                if not self.name:
                    # XXX Check password
                    self.name = args[0]
                    self.write('Welcome to the fray, %s.' % self.name)
                    self.manager.enter_lobby(self)
                else:
                    self.err('Already logged in.')
            elif cmd == '^':
                # Send to manager
                ret = self.manager.player_cmd(args)
                self.write(ret)
            elif not self.name:
                self.err('Log in first.')
            else:
                # Send to game
                if not self.game:
                    self.pending = (cmd, args)
                    self.block()
                else:
                    self.game.handle(self, cmd, args)
        except Exception as err:
            traceback.print_exc()
            self.err(str(err))


class Game:
    def __init__(self, manager, players):
        self.manager = manager
        self.players = players
        self.setup()

    def declare_winner(self, player):
        self.manager.declare_winner(self, player)


def run(nplayers, game_factory, port, auth):
    flagger = Flagger(('localhost', 6668), auth)
    manager = Manager(2, game_factory, flagger)
    listener = Listener(('', port), Player, manager)
    asyncore.loop()

