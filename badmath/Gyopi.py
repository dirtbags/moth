import badmath
import time
import os
import traceback
import pickle
from hashlib import sha256

try:
    from ctf import irc
    from ctf.flagd import Flagger
except:
    import sys
    sys.path.append('/home/pflarr/repos/gctf/')
    from ctf.flagd import Flagger
    from ctf import irc

class Gyopi(irc.Bot):
    STATE_FN = 'badmath.state'

    SALT = b'this is questionable.'

    FLAG_DEFAULT = 'dirtbags'
    MAX_ATTEMPT_RATE = 3
    NOBODY = '\002[nobody]\002'

    FLAG_HOST = b'ctf1.lanl.gov'
#    FLAG_HOST = b'localhost'

    def __init__(self, host, channels, dataPath, flagger):
        irc.Bot.__init__(self, host, ['gyopi'], 'Gyopi', channels)

        self._dataPath = dataPath

        self._flag = flagger

        try:
            self._loadState()
        except:
            traceback.print_exc()
            self._lvl = 0
            self._flag.set_flag( self.FLAG_DEFAULT )

            self._tokens = []
            self._lastAttempt = {}
            self._affiliations = {}
            self._newPuzzle()

    def err(self, exception):
        """Save the traceback for later inspection"""
        irc.Bot.err(self, exception)
        t,v,tb = exception
        info = []
        while 1:
            info.append('%s:%d(%s)' % 
                        (os.path.basename(tb.tb_frame.f_code.co_filename),
                                          tb.tb_lineno,
                                          tb.tb_frame.f_code.co_name))
            tb = tb.tb_next
            if not tb:
                break
        del tb                          # just to be safe
        infostr = '[' + '] ['.join(info) + ']'
        self.last_tb = '%s %s %s' % (t, v, infostr)
        print(self.last_tb)

    def cmd_JOIN(self, sender, forum, addl):
        """On join, announce who has the flag."""
        if sender.name() in self.nicks:
            self._tellFlag(forum)
            self._tellPuzzle(forum)
    
    def _newPuzzle(self):
        """Create a new puzzle."""
        self._key, self._puzzle, self._banned = badmath.mkPuzzle(self._lvl)

    def _loadState(self):
        """Load the last state from the stateFile."""
        statePath = os.path.join(self._dataPath, self.STATE_FN)
        stateFile = open( statePath, 'br' )
        state = pickle.load(stateFile)
        self._lvl = state['lvl']
        self._flag.set_flag( state['flag'] )
        self._lastAttempt = state['lastAttempt']
        self._affiliations = state['affiliations']
        self._puzzle = state['puzzle']
        self._key = state['key']
        self._banned = state['banned']
        self._tokens = state.get('tokens', [])

    def _saveState(self):
        """Write the current state to file."""
        state = {'lvl': self._lvl,
                 'flag': self._flag.flag,
                 'lastAttempt': self._lastAttempt,
                 'affiliations': self._affiliations,
                 'puzzle': self._puzzle,
                 'key': self._key,
                 'banned': self._banned,
                 'tokens': self._tokens}

        # Do the write as an atomic move operation
        statePath = os.path.join(self._dataPath, self.STATE_FN)
        stateFile = open(statePath + '.tmp', 'wb')
        pickle.dump(state, stateFile)
        stateFile.close()
        os.rename( statePath + '.tmp', statePath)

    def _tellFlag(self, forum):
        """Announce who owns the flag."""
        forum.msg('%s has the flag.' % (self._flag.flag))

    def _tellPuzzle(self, forum):
        """Announce the current puzzle."""
        forum.msg('Difficulty level is %d' % self._lvl)
        forum.msg('The problem is: %s' % ' '.join( map(str, self._puzzle)))

    def _getStations(self):
        stations = {}
        with open(os.path.join(STORAGE, 'stations.txt')) as file:
            lines = file.readlines()
            for line in lines:
                try:
                    name, file = line.split(':')
                except:
                    continue
                stations[name] = file

        return stations

    def _giveToken(self, user, forum):
        """Hand a Jukebox token to the user."""

        token = self._jukebox.mkToken(user)

        forum.msg('You get a jukebox token: %s' % token)
        forum.msg('Use this with the !set command to change the music.')
        forum.msg('This token is specific to your user name, and is only '
                  'useable once.')

    def _useToken(self, user, forum, token, station):
        """Use the given token, and change the current station to station."""
        try:
            station = int(station)
            stations = self._getStations()
            assert station in stations
        except:
            forum.msg('%s: Invalid Station (%s)' % station)
            return
            
        if token in self._tokens[user]:
            self._tokens[user].remove(token)


    def cmd_PRIVMSG(self, sender, forum, addl):
        text = addl[0]
        who = sender.name()
        if text.startswith('!'):
            parts = text[1:].split(' ', 1)
            cmd = parts[0]
            if len(parts) > 1:
                args = parts[1]
            else:
                args = None
            if cmd.startswith('r'):
                # Register
                if args:
                    self._affiliations[who] = args
                team = self._affiliations.get(who, self.NOBODY)
                forum.msg('%s is playing for %s' % (who, team))
            elif cmd.startswith('w'):
                forum.msg('Teams:')
                for player in self._affiliations:
                    forum.msg('%s: %s' % (player, self._affiliations[player]))
            elif cmd.startswith('embrace'):
                # Embrace
                forum.ctcp('ACTION', 'is devoid of emotion.')
            elif cmd.startswith('f'):
                # Flag
                self._tellFlag(forum)
            elif cmd.startswith('h'):
                # Help
                forum.msg('''Goal: Help me with my math homework, FROM ANOTHER DIMENSION!  Order of operations is always left to right in that dimension, but the operators are alien.''')
                forum.msg('Order of operations is always left to right.')
                #forum.msg('Goal: The current winner gets to control the contest music.')
                forum.msg('Commands: !help, !flag, !register [TEAM], !solve SOLUTION,!? EQUATION, !ops, !problem, !who')
            elif cmd.startswith('prob'):
                self._tellPuzzle(forum)
            elif cmd.startswith('solve') and args:
                # Solve
                team = self._affiliations.get(who)
                lastAttempt = time.time() - self._lastAttempt.get(team, 0)
                #UN-COMMENT AFTER NMT CTF
#                self._lastAttempt[team] = time.time()
                answer = badmath.solve(self._key, self._puzzle)
                try:
                    attempt = int(''.join(args).strip())
                except:
                    forum.msg("%s: Answers are always integers.")
                if not team:
                    forum.msg('%s: register first (!register TEAM).' % who)
                elif self._flag.flag == team:
                    forum.msg('%s: Greedy, greedy.' % who)
                elif lastAttempt < self.MAX_ATTEMPT_RATE:
                    forum.msg('%s: Wait at least %d seconds between attempts' %
                              (team, self.MAX_ATTEMPT_RATE))
                elif answer == attempt:
                    self._flag.set_flag( team )
                    self._lvl = self._lvl + 1
                    self._tellFlag(forum)
                    self._newPuzzle()
                    self._tellPuzzle(forum)
#                    self._giveToken(who, sender)
                    self._saveState()
                else:
                    forum.msg('%s: That is not correct.' % who)

            # Test a simple one op command.
            elif cmd.startswith('?'):
                if not args:
                    forum.msg('%s: Give me an easier problem, and I\'ll '
                              'give you the answer.' % who)
                    return 

                try:
                    tokens = badmath.parse(''.join(args))
                except (ValueError) as msg:
                    forum.msg('%s: %s' % (who, msg))
                    return 

                if len(tokens) > 3:
                    forum.msg('%s: You can only test one op at a time.' % who)
                    return

                for num in self._banned:
                    if num in tokens:
                        forum.msg('%s: You can\'t test numbers in the '
                                  'puzzle.' % who)
                        return
                
                try:
                    result = badmath.solve(self._key, tokens)
                    forum.msg('%s: %s -> %d' % (who, ''.join(args), result)) 
                except Exception as msg:
                    forum.msg("%s: That doesn't work at all: %s" % (who, msg))

            elif cmd == 'birdzerk':
                self._saveState()

            elif cmd == 'traceback':
                forum.msg(self.last_tb or 'No traceback')

if __name__ == '__main__':
    import optparse

    p = optparse.OptionParser()
    p.add_option('-i', '--irc', dest='ircHost', default='localhost',
                 help='IRC Host to connect to.')
    p.add_option('-f', '--flagd', dest='flagd', default='localhost',
                 help='Flag Server to connect to')
    p.add_option('-p', '--password', dest='password',
                 default='badmath:::a41c6753210c0bdafd84b3b62d7d1666',
                 help='Flag server password')
    p.add_option('-d', '--path', dest='path', default='/var/lib/badmath',
                 help='Path to where we can store state info.')
    p.add_option('-c', '--channel', dest='channel', default='#badmath',
                 help='Which channel to join')

    opts, args = p.parse_args()
    channels = [opts.channel]

    flagger = Flagger(opts.flagd, opts.password.encode('utf-8')) 
    gyopi = Gyopi((opts.ircHost, 6667), channels, opts.path, flagger)
    irc.run_forever()
