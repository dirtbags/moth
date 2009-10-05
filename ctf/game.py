#! /usr/bin/env python3

import json
import asyncore
import asynchat
import socket
import traceback
import time
from errno import EPIPE
from . import teams
from . import Flagger


# Heartbeat frequency (in seconds)
pulse = 2.0

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

    def readable(self):
        self.manager.heartbeat(time.time())
        return True


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

    def __init__(self, game_factory, flagger, minplayers, maxplayers=None):
        self.game_factory = game_factory
        self.flagger = flagger
        self.minplayers = minplayers
        self.maxplayers = maxplayers or minplayers
        self.games = set()
        self.lobby = set()
        self.contestants = []
        self.last_beat = 0
        self.timers = set()

    def heartbeat(self, now):
        """Called by listener to beat heart."""

        now = time.time()
        if now > self.last_beat + pulse:
            for game in list(self.games):
                game.heartbeat(now)
            self.last_beat = now
        for event in self.timers:
            when, cb = event
            if now >= when:
                self.timers.remove(event)
                cb()

    def add_timer(self, when, cb):
        """Add a timed callback."""

        self.timers.add((when, cb))

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
        """Start a new contest."""

        self.contestants = list(self.lobby)
        print('new playoff:', [c.name for c in self.contestants])

    def run_contest(self):
        # Purge any disconnected players
        self.contestants = [p for p in self.contestants if p.connected]
        self.lobby = set([p for p in self.lobby if p.connected])

        llen = len(self.lobby)
        clen = len(self.contestants)
        glen = len(self.games)
        if llen == 1:
            # Give the flag to the only team connected
            self.set_flag(list(self.lobby)[0])
        elif llen < self.minplayers:
            # More than one connected team, but still not enough to play
            self.set_flag(None)
        elif (clen == 1) and (glen == 0):
            # Give the flag to the last team standing, and start a new contest
            self.set_flag(self.contestants.pop())
            self.start_contest()
        elif (llen >= self.minplayers) and (clen == 0) and (glen == 0):
            # There are enough in the lobby to begin a contest now
            self.start_contest()

        while len(self.contestants) >= self.minplayers:
            players = self.contestants[:self.maxplayers]
            del self.contestants[:self.maxplayers]
            game = self.game_factory(self, set(players))
            self.games.add(game)
            for player in players:
                player.attach_game(game)

    def declare_winner(self, game, winner=None):
        print('Winner:', winner and winner.name)
        self.games.remove(game)

        # Winner stays in the contest
        if winner:
            self.add_contestant(winner)

    def player_cmd(self, args):
        cmd = args[0].lower()
        if cmd == 'lobby':
            return [p.name for p in self.lobby]
        elif cmd == 'games':
            return len(self.games)
        elif cmd == 'flag':
            return self.flagger.flag
        else:
            raise ValueError('Unrecognized manager command')


class Player(asynchat.async_chat):
    # How long can a connection not send anything at all (unless blocked)?
    timeout = 10.0

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
        ret = (not self.blocked) and asynchat.async_chat.readable(self)
        if ret:
            if time.time() - self.last_activity > self.timeout:
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
                if self.name:
                    self.err('Already logged in.')
                elif teams.chkpasswd(args[0], args[1]):
                    self.name = args[0]
                    self.write('Welcome to the fray, %s.' % self.name)
                    self.manager.enter_lobby(self)
                else:
                    self.err('Invalid password.')
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
        self.unblock()
        if self.game:
            self.game.player_died(self)
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

    def setup(self):
        pass

    def heartbeat(self, now):
        pass

    def declare_winner(self, winner):
        self.manager.declare_winner(self, winner)

        # Congratulate winner
        if winner:
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
        """Player forfeits the game."""

        self.remove(player)

    def remove(self, player):
        """Remove the player from the game."""

        self.players.remove(player)
        player.detach_game()

    def player_died(self, player):
        self.forfeit(player)


class TurnBasedGame(Game):
    # How long you get to make a move (in seconds)
    move_timeout = 2.0

    # How long you get to complete the game (in seconds)
    game_timeout = 6.0

    def __init__(self, manager, players):
        now = time.time()
        self.ended_turn = set()
        self.running = True
        self.winner = None
        self.lastmoved = dict([(p, now) for p in players])
        self.began = now
        Game.__init__(self, manager, players)

    def heartbeat(self, now=None):
        print('heart', self)
        if now and (now - self.began > self.game_timeout):
            self.running = False

        # Idle players forfeit.  They're also booted, so we don't have
        # to worry about the synchronous illusion.
        for player in list(self.players):
            if not player.connected:
                self.remove(player)
                continue
            when = self.lastmoved[player]
            if now - when > self.move_timeout:
                player.err('Timeout waiting for a move')
                player.close()

        # If everyone left, nobody wins.
        if not self.players:
            self.manager.declare_winner(self, None)

    def player_died(self, player):
        Game.player_died(self, player)
        if player in self.players:
            # Update stuff
            self.heartbeat()

    def declare_winner(self, winner):
        """Declare winner.

        In a turn-based game, you can't tell anyone that the game has
        ended until they make a move.  Otherwise, you ruin the illusion
        of the game being synchronous.  This only sets the winner variable,
        which is checked in self.end_turn().

        """

        self.running = False
        self.winner = winner

    def calculate_moves(self):
        """Calculate all moves at the end of a turn.

        Override this to define what to do when every player has ended
        their turn.

        """
        pass

    def end_turn(self, player):
        """End player's turn."""

        now = time.time()

        self.ended_turn.add(player)
        self.lastmoved[player] = now
        if not self.players:
            self.manager.declare_winner(self, None)
        elif len(self.players) == 1:
            winners = list(self.players)
            self.declare_winner(winners[0])
        elif len(self.ended_turn) >= len(self.players):
            self.calculate_moves()
            if self.running:
                for p in self.players:
                    p.unblock()
            else:
                # Game has ended, tell everyone how they did
                for p in list(self.players):
                    if self.winner == p:
                        p.win()
                    else:
                        p.lose()
                self.manager.declare_winner(self, self.winner)
            self.ended_turn = set()
        elif self.running:
            player.block()
        else:
            # The game has ended, tell the player, now that they've made
            # a move.
            if self.winner == player:
                player.win()
                self.manager.declare_winner(self, self.winner)
            else:
                player.lose()
            self.remove(player)


##
## Running a game
##

def start(game_factory, port, auth, minplayers, maxplayers=None):
    flagger = Flagger(('localhost', 6668), auth)
    manager = Manager(game_factory, flagger, minplayers, maxplayers)
    listener = Listener(('', port), Player, manager)
    return (flagger, manager, listener)

def run(game_factory, port, auth, minplayers, maxplayers=None):
    start(game_factory, port, auth, minplayers, maxplayers)
    asyncore.loop(use_poll=True)

