#! /usr/bin/env python3

import json
import asyncore
import asynchat
import socket
import traceback
import time
from errno import EPIPE


# Number of seconds (roughly) you can be idle before you pass your turn
timeout = 30.0

# The current time of day
now = time.time()

# Heartbeat frequency (in seconds)
pulse = 2.0

##
## Heartbeat stuff
##

hearts = set()
last_beat = 0

def add_heart(cb):
    global hearts

    hearts.add(cb)


def del_heart(cb):
    global hearts

    hearts.remove(cb)


def beat_heart():
    global hearts, last_beat, now

    if now - last_beat > pulse:
        last_beat = now
        for cb in hearts:
            try:
                cb()
            except:
                traceback.print_exc()


##
## Network stuff
##

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
        self.push(auth + b'\n')
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
        self.push(team.encode('utf-8') + b'\n')
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
        self.lobby = set()
        self.contestants = []
        add_heart(self.heartbeat)

    def heartbeat(self):
        games = list(self.games)
        for game in games:
            print('heartbeat', game)
            game.heartbeat()

    def enter_lobby(self, player):
        self.lobby.add(player)
        self.run_contest()

    def add_contestant(self, player):
        self.contestants.append(player)
        self.run_contest()

    def disconnect(self, player):
        """Player has disconnected."""

        pass

    def set_flag(self, player):
        """Player has the flag."""

        self.flagger.set_flag(player.name)

    def start_contest(self):
        """Start a new contest.

        This is where we purge any disconnected clients from the lobby.
        """

        self.contestants = []
        gone = set()
        for player in self.lobby:
            if player.connected:
                self.contestants.append(player)
            else:
                gone.add(player)
        self.lobby.difference_update(gone)

    def run_contest(self):
        # Purge any disconnected players
        self.contestants = [p for p in self.contestants if p.connected]
        self.lobby = set([p for p in self.lobby if p.connected])

        # This is the closest thing we get to pattern matching in python
        llen = len(self.lobby)
        clen = len(self.contestants)
        glen = len(self.games)
        if   (((llen == 1)                                                       )):
            # Give the flag to the only team connected
            self.set_flag(list(self.lobby)[0])
        elif ((                            (clen == 1)             and (glen == 0))):
            # Give the flag to the last team standing, and start a new contest
            self.set_flag(self.contestants.pop())
            self.start_contest()
        if   (((llen == 0)             and (clen == 0)             and (glen == 0)) or
              ((llen < self.nplayers)  and (clen == 0)             and (glen == 0)) or
              (                            (clen < self.nplayers)  and (glen >= 1))):
            pass
        elif (((llen >= self.nplayers) and (clen == 0)             and (glen == 0))):
            self.start_contest()

        while len(self.contestants) >= self.nplayers:
            players = self.contestants[:self.nplayers]
            del self.contestants[:self.nplayers]
            game = self.game_factory(self, players)
            self.games[game] = players
            for player in players:
                player.attach_game(game)

    def declare_winner(self, game, winner):
        print('winner', game)
        players = self.games[game]
        del self.games[game]

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
        self.last_activity = time.time()

    def readable(self):
        global now, timeout

        ret = (not self.blocked) and asynchat.async_chat.readable(self)
        if ret:
            if now - self.last_activity > timeout:
                # They waited too long.
                self.err('idle timeout')
                self.close()
                return False
        return ret

    def block(self):
        """Block reads"""
        self.blocked = True

    def unblock(self):
        """Unblock reads"""
        self.blocked = False
        self.last_activity = time.time()

    def attach_game(self, game):
        self.game = game
        if self.pending:
            self.unblock()
            self.game.handle(self, *self.pending)
            self.pending = None

    def detach_game(self):
        self.game = None

    def _write_val(self, val):
        s = json.dumps(val) + '\n'
        self.push(s.encode('utf-8'))

    def write(self, val):
        self._write_val(['OK', val])

    def err(self, msg):
        self._write_val(['ERR', msg])

    def win(self):
        self.detach_game()
        self._write_val(['WIN'])
        self.unblock()

    def lose(self):
        self.detach_game()
        self._write_val(['LOSE'])
        self.unblock()

    def collect_incoming_data(self, data):
        self.inbuf.append(data)
        if len(self.inbuf) > 10:
            self.err('Too much data, punk.')
            self.close()

    def found_terminator(self):
        self.last_activity = time.time()
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

    def close(self):
        if self.game:
            self.game.disconnect(self)
        self.manager.disconnect(self)
        asynchat.async_chat.close(self)

    def send(self, data):
        try:
            return asynchat.async_chat.send(self, data)
        except socket.error as why:
            if why.args[0] == EPIPE:
                # Broken pipe, shut down.
                self.close()
            else:
                raise


class Game:
    def __init__(self, manager, players):
        self.manager = manager
        self.players = players
        self.setup()

    def heartbeat(self):
        pass

    def declare_winner(self, winner):
        self.manager.declare_winner(self, winner)

        # Congratulate winner
        winner.win()

        # Inform losers of their loss
        losers = [p for p in players if p != winner]
        for p in losers:
            p.lose()


    def handle(self, player, cmd, args):
        """Handle a command from player.

        This just dispatches to 'self.do_[cmd]'.

        """

        method_name = 'do_%s' % cmd
        try:
            method = getattr(self, method_name)
            method(player, args)
        except AttributeError:
            raise ValueError('Invalid command: %s' % cmd)

    def forfeit(self, player):
        """Player forfeits the game, in a 2-player game.

        If your game has more than 2 players, you need to define
        your own forfeit method.

        """

        if len(self.players) == 2:
            if player == self.players[0]:
                self.declare_winner(self.players[1])
            else:
                self.declare_winner(self.players[0])
        else:
            raise NotImplementedError('forfeit method undefined')

    def disconnect(self, player):
        """Disconnect the player."""

        self.forfeit(player)


class TurnBasedGame(Game):
    # How long you get to make a move (in seconds)
    move_timeout = 2.0

    def __init__(self, manager, players):
        global now

        self.ended_turn = set()
        self.winner = None
        self.lastmoved = dict([(p, now) for p in players])
        Game.__init__(self, manager, players)

    def heartbeat(self):
        global now

        for p, when in self.lastmoved.items():
            if now - when > self.move_timeout:
                self.disconnect(p)
            if self.winner:
                break

    def disconnect(self, player):
        Game.disconnect(self, player)
        self.end_turn(player)

    def declare_winner(self, winner):
        """Declare winner.

        In a turn-based game, you can't tell anyone that the game has
        ended until they make a move.  Otherwise, you ruin the illusion
        of the game being synchronous.  This only sets the winner variable,
        which is checked in self.end_turn().

        """

        self.manager.declare_winner(self, winner)
        self.winner = winner

    def calculate_moves(self):
        """Calculate all moves at the end of a turn.

        Override this to define what to do when every player has ended
        their turn.

        """
        pass

    def end_turn(self, player):
        """End player's turn."""

        global now

        # The player has ended their turn; it's okay to tell them now
        # that the game has ended.
        if self.winner:
            if self.winner == player:
                player.win()
            else:
                player.lose()
            return

        self.ended_turn.add(player)
        self.lastmoved[player] = now
        player.block()
        if len(self.ended_turn) == len(self.players):
            for p in self.players:
                p.unblock()
            self.calculate_moves()
            self.ended_turn = set()



##
## Running a game
##

def loop():
    global timeout, pulse, now

    my_timeout = min(timeout, pulse)

    while True:
        now = time.time()
        beat_heart()
        asyncore.poll2(timeout=my_timeout)


def run(nplayers, game_factory, port, auth):
    flagger = Flagger(('localhost', 6668), auth)
    manager = Manager(2, game_factory, flagger)
    listener = Listener(('', port), Player, manager)
    loop()

