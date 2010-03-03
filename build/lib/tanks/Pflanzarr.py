import fcntl
import math
import os
import random
import cgi
from sets import Set as set
from ctf import teams, html, paths
from cStringIO import StringIO

from urllib import unquote, quote

import Tank

class NotEnoughPlayers(Exception):
    pass

class Pflanzarr:
    SPACING = 150

    def __init__(self, dir):
        """Initialize a new game of Pflanzarr.
@param dir: The data directory."""

        # Setup the game environment
        self._setupDirectories(dir)

        # Figure out what game number this is.
        self.gameNum = self._getGameNum()
        self.gameFilename = os.path.join(self._resultsDir, '%04d.html' % self.gameNum)

        tmpPlayers = os.listdir(self._playerDir)
        players = []
        for p in tmpPlayers:
            p = unquote(p)
            if (not (p.startswith('.')
                     or p.endswith('#')
                     or p.endswith('~'))
                and teams.exists(p)):
                players.append(p)

        AIs = {}
        for player in players:
            AIs[player] = open(os.path.join(self._playerDir, player)).read()
        defaultAIs = self._getDefaultAIs(dir)

        if len(players) < 1:
            raise NotEnoughPlayers()

        # The one is added to ensure that there is at least one house
        # bot.
        cols = math.sqrt(len(players) + 1)
        if int(cols) != cols:
            cols = cols + 1

        cols = int(cols)
        cols = max(cols, 2)

        rows = len(players)/cols
        if len(players) % cols != 0:
            rows = rows + 1
        rows = max(rows, 2)

        self._board = (cols*self.SPACING, rows*self.SPACING)

        while len(players) < cols*rows:
            players.append(None)

        self._tanks = []
        for i in range(cols):
            for j in range(rows):
                startX = i*self.SPACING + self.SPACING/2
                startY = j*self.SPACING + self.SPACING/2
                player = random.choice(players)
                players.remove(player)
                color = '#' + teams.color(player)
                tank = Tank.Tank( player, (startX, startY), color,
                                  self._board, testMode=True)
                if player == None:
                    tank.program(random.choice(defaultAIs))
                else:
                    tank.program(AIs[player])
                self._tanks.append(tank)

        # We only want to make these once, so we do it here.
        self._tanksByX = list(self._tanks)
        self._tanksByY = list(self._tanks)

        self._deadTanks = set()

    def run(self, maxTurns=None):
        kills = {}
        for tank in self._tanks:
            kills[tank] = set()

        # Open HTML output
        hdr = StringIO()
        hdr.write('<script type="application/javascript" src="../tanks.js"></script>\n'
                  '<script type="application/javascript">\n')
        hdr.write('turns = [\n')

        turn = 0
        lastTurns = 3
        while ((maxTurns is None) or turn < maxTurns) and lastTurns > 0:
            if len(self._tanks) - len(self._deadTanks) < 2:
                lastTurns = lastTurns - 1

            near = self._getNear()
            deadThisTurn = set()

            liveTanks = set(self._tanks).difference(self._deadTanks)

            for tank in liveTanks:
                # Shoot now, if we said to shoot last turn
                dead = tank.fire( near[tank] )
                kills[tank] = kills[tank].union(dead)
                self._killTanks(dead, 'Shot by %s' % cgi.escape(tank.name or teams.house))

            for tank in liveTanks:
                # We also check for collisions here, while we're at it.
                dead = tank.sense( near[tank] )
                kills[tank] = kills[tank].union(dead)
                self._killTanks(dead, 'Collision')

            hdr.write(' [\n')

            # Draw the explosions
            for tank in self._deadTanks:
                tank.draw(hdr)

            # Draw the live tanks.
            for tank in self._tanksByX:
                # Have the tank run its program, move, etc.
                tank.draw(hdr)

            hdr.write(' ],\n')

            # Have the live tanks do their turns
            for tank in self._tanksByX:
                tank.execute()

            turn = turn + 1

        # Removes tanks from their own kill lists.
        for tank in kills:
            if tank in kills[tank]:
                kills[tank].remove(tank)

        for tank in self._tanks:
            self._outputErrors(tank)

        hdr.write('];\n')
        hdr.write('</script>\n')

        # Decide on the winner
        winner = self._chooseWinner(kills)
        self.winner = winner.name

        # Now generate HTML body
        body = StringIO()
        body.write('    <canvas id="battlefield" width="%d" height="%d">\n' % self._board)
        body.write('      Sorry, you need an HTML5-capable browser to see this.\n'
                   '    </canvas>\n'
                   '    <p>\n')
        if self.gameNum > 0:
            body.write('      <a href="%04d.html">&larr; Prev</a> |' %
                       (self.gameNum - 1))
        body.write('      <a href="%04d.html">Next &rarr;</a> |' %
                   (self.gameNum + 1))
        body.write('      <span id="fps">0</span> fps\n'
                   '    </p>\n'
                   '    <table class="results">\n'
                   '      <tr>\n'
                   '        <th>Team</th>\n'
                   '        <th>Kills</th>\n'
                   '        <th>Cause of Death</th>\n'
                   '      </tr>\n')

        tanks = self._tanks[:]
        tanks.remove(winner)
        tanks[0:0] = [winner]
        for tank in tanks:
            if tank is winner:
                rowStyle = ('style="font-weight: bold; '
                            'color: #000; '
                            'background-color: %s;"' % tank.color)
            else:
                rowStyle = 'style="background-color:%s; color: #000;"' % tank.color
            if tank.name:
                name = cgi.escape(tank.name)
            else:
                name = teams.house
            body.write('<tr %s><td>%s</td><td>%d</td><td>%s</td></tr>' %
                       (rowStyle,
                        name,
                        len(kills[tank]),
                        cgi.escape(tank.deathReason)))
        body.write('  </table>\n')

        # Write everything out
        html.write(self.gameFilename,
                   'Tanks round %d' % self.gameNum,
                   body.getvalue(),
                   hdr=hdr.getvalue(),
                   onload='start(turns);')



    def _killTanks(self, tanks, reason):
        for tank in tanks:
            if tank in self._tanksByX:
                self._tanksByX.remove(tank)
            if tank in self._tanksByY:
                self._tanksByY.remove(tank)

            tank.die(reason)

        self._deadTanks = self._deadTanks.union(tanks)

    def _chooseWinner(self, kills):
        """Choose a winner.  In case of a tie, live tanks prevail, in case
    of further ties, a winner is chosen at random.  This outputs the winner
    to the winners file and outputs a results table html file."""
        tanks = list(self._tanks)
        def winSort(t1, t2):
            """Sort by # of kill first, then by life status."""
            result = cmp(len(kills[t1]), len(kills[t2]))
            if result != 0:
                return result

            if t1.isDead and not t2.isDead:
                return -1
            elif not t1.isDead and t2.isDead:
                return 1
            else:
                return 0
        tanks.sort(winSort)
        tanks.reverse()

        # Get the list of potential winners
        winners = []
        for i in range(len(tanks)):
            if len( kills[tanks[0]] ) == len( kills[tanks[i]] ) and \
               tanks[0].isDead == tanks[i].isDead:
                winners.append(tanks[i])
            else:
                break
        winner = random.choice(winners)
        return winner


    def _outputErrors(self, tank):
        """Output errors for each team."""
        if tank.name == None:
            return

        if tank._program.errors:
            print tank.name, 'has errors'


        fileName = os.path.join(self._errorDir, quote(tank.name, ''))
        file = open(fileName, 'w')
        for error in tank._program.errors:
            file.write(error)
            file.write('\n')
        file.close()

    def _getNear(self):
        """A dictionary of the set of tanks nearby each tank.  Nearby is
    defined as within the square centered the tank with side length equal
    twice the sensor range.  Only a few tanks within the set (those in the
    corners of the square) should be outside the sensor range."""

        self._tanksByX.sort(lambda t1, t2: cmp(t1.pos[0], t2.pos[0]))
        self._tanksByY.sort(lambda t1, t2: cmp(t1.pos[1], t2.pos[1]))

        nearX = {}
        nearY = {}
        for tank in self._tanksByX:
            nearX[tank] = set()
            nearY[tank] = set()

        numTanks = len(self._tanksByX)
        offset = 1
        for index in range(numTanks):
            cTank = self._tanksByX[index]
            maxRange = cTank.SENSOR_RANGE + cTank.RADIUS + 1
            near = set([cTank])
            for i in [(j + index) % numTanks for j in range(1, offset)]:
                near.add(self._tanksByX[i])
            while offset < numTanks:
                nTank = self._tanksByX[(index + offset) % numTanks]
                if (index + offset >= numTanks and
                    self._board[0] + nTank.pos[0] - cTank.pos[0] < maxRange):
                        near.add(nTank)
                        offset = offset + 1
                elif (index + offset < numTanks and
                      nTank.pos[0] - cTank.pos[0] < maxRange ):
                    near.add(nTank)
                    offset = offset + 1
                else:
                    break

            if offset > 1:
                offset = offset - 1

            for tank in near:
                nearX[tank] = nearX[tank].union(near)

        offset = 1
        for index in range(numTanks):
            cTank = self._tanksByY[index]
            maxRange = cTank.SENSOR_RANGE + cTank.RADIUS + 1
            near = set([cTank])
            for i in [(j + index) % numTanks for j in range(1, offset)]:
                near.add(self._tanksByY[i])
            while offset < numTanks:
                nTank = self._tanksByY[(index + offset) % numTanks]
                if (index + offset < numTanks and
                    nTank.pos[1] - cTank.pos[1] < maxRange):
                    near.add(nTank)
                    offset = offset + 1
                elif (index + offset >= numTanks and
                      self._board[1] + nTank.pos[1] - cTank.pos[1] < maxRange):
                    near.add(nTank)
                    offset = offset + 1
                else:
                    break

            if offset > 1:
                offset = offset - 1

            for tank in near:
                nearY[tank] = nearY[tank].union(near)

        near = {}
        for tank in self._tanksByX:
            near[tank] = nearX[tank].intersection(nearY[tank])
            near[tank].remove(tank)

        return near

    def _setupDirectories(self, dir):
        """Setup all the directories needed by the game."""

        if not os.path.exists(dir):
            os.mkdir(dir)

        self._dir = dir

        # Don't run more than one game at the same time.
        self._lockFile = open(os.path.join(dir, '.lock'), 'a')
        try:
            fcntl.flock(self._lockFile, fcntl.LOCK_EX|fcntl.LOCK_NB)
        except:
            sys.exit(1)

        # Setup all the directories we'll need.
        self._resultsDir = os.path.join(dir, 'results')
        self._errorDir = os.path.join(dir, 'errors')
        self._playerDir = os.path.join(dir, 'ai', 'players')

    def _getDefaultAIs(self, basedir):
        """Load all the house bot AIs."""
        defaultAIs = []

        path = os.path.join(basedir, 'ai', 'house')
        files = os.listdir(path)
        for fn in files:
            if fn.startswith('.'):
                continue

            fn = os.path.join(path, fn)
            file = open(fn)
            defaultAIs.append(file.read())

        return defaultAIs

    def _getGameNum(self):
        """Figure out what game number this is from the past games played."""

        games = os.listdir(self._resultsDir)
        games.sort()
        if games:
            fn = games[-1]
            s, _ = os.path.splitext(fn)
            return int(s) + 1
        else:
            return 0

if __name__ == '__main__':
    import sys, traceback
    try:
        p = Pflanzarr(sys.argv[1])
        p.run(int(sys.argv[3]))
    except:
        traceback.print_exc()
        print "Usage: Pflanzarr.py dataDirectory #turns"


