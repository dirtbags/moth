import fcntl
import math
import os
import random
import subprocess
import xml.sax.saxutils

from urllib import unquote, quote

from PIL import Image, ImageColor, ImageDraw

import Tank

class Pflanzarr:

    TEAMS_FILE = '/var/lib/ctf/passwd'

    FRAME_DELAY = 15

    SPACING = 150
    backgroundColor = '#ffffff'

    def __init__(self, dir, difficulty='easy'):
        """Initialize a new game of Pflanzarr.
@param dir: The data directory."""

        assert difficulty in ('easy', 'medium', 'hard')

        # Setup the game environment
        self._setupDirectories(dir)

        # Figure out what game number this is.
        self._gameNum = self._getGameNum()
        self._gameDir = os.path.join(self._resultsDir, str(self._gameNum))
        if not os.path.exists(self._gameDir):
            os.mkdir(self._gameDir)

        colors = self._getColors()

        tmpPlayers = os.listdir(self._playerDir)
        players = []
        for p in tmpPlayers:
            p = unquote(p)
            if not (p.startswith('.') or p.endswith('#') or p.endswith('~'))\
               and p in colors:
                players.append(p)

        AIs = {}
        for player in players:
            AIs[player] = open(os.path.join(self._playerDir, player)).read()
        defaultAIs = self._getDefaultAIs(dir, difficulty)

        assert len(players) >= 1, "There must be at least one player."

        # The one is added to ensure that there is at least one #default bot.
        cols = math.sqrt(len(players) + 1)
        if int(cols) != cols:
            cols = cols + 1

        cols = int(cols)
        if cols < 2:
            cols = 2

        rows = len(players)/cols
        if len(players) % cols != 0:
            rows = rows + 1

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
                if player == None:
                    color = '#a0a0a0'
                else:
                    color = colors[player]
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
    
        print "Starting new game with %d players." % len(self._tanks)

        kills = {}
        for tank in self._tanks:
            kills[tank] = set()

        turn = 0
        lastTurns = 3
        while ((maxTurns is None) or turn < maxTurns) and lastTurns > 0: 
            if len(self._tanks) - len(self._deadTanks) < 2:
                lastTurns = lastTurns - 1

            image = Image.new('RGB', self._board)
            draw = ImageDraw.Draw(image)
            draw.rectangle(((0,0), self._board), fill=self.backgroundColor)
            near = self._getNear()
            deadThisTurn = set()

            liveTanks = set(self._tanks).difference(self._deadTanks)

            for tank in liveTanks:
                # Shoot now, if we said to shoot last turn
                dead = tank.fire( near[tank] ) 
                kills[tank] = kills[tank].union(dead)
                self._killTanks(dead, 'Shot by %s' % tank.name)

            for tank in liveTanks:
                # We also check for collisions here, while we're at it.
                dead = tank.sense( near[tank] ) 
                kills[tank] = kills[tank].union(dead)
                self._killTanks(dead, 'Collision')

            # Draw the explosions
            for tank in self._deadTanks:
                tank.draw(image)

            # Draw the live tanks.
            for tank in self._tanksByX:
                # Have the tank run its program, move, etc.
                tank.draw(image)

            # Have the live tanks do their turns
            for tank in self._tanksByX:
                tank.execute()
            
            fileName = os.path.join(self._imageDir, '%05d.ppm' % turn)
            image.save(fileName, 'PPM')
            turn = turn + 1

        # Removes tanks from their own kill lists.
        for tank in kills:
            if tank in kills[tank]:
                kills[tank].remove(tank)

        for tank in self._tanks:
            self._outputErrors(tank)
        self._makeMovie()
        # This needs to go after _makeMovie; the web scripts look for these
        # files to see if the game has completed.
        self._saveResults(kills)

    def _killTanks(self, tanks, reason):
        for tank in tanks:
            if tank in self._tanksByX:
                self._tanksByX.remove(tank)
            if tank in self._tanksByY:
                self._tanksByY.remove(tank)

            tank.die(reason)

        self._deadTanks = self._deadTanks.union(tanks)

    def _saveResults(self, kills):
        """Choose a winner.  In case of a tie, live tanks prevail, in case
    of further ties, a winner is chosen at random.  This outputs the winner
    to the winners file and outputs a results table html file."""
        resultsFile = open(os.path.join(self._gameDir, 'results.html'), 'w')
        winnerFile = open(os.path.join(self._dir, 'winner'),'w')

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
        tanks.sort(winSort, reverse=1)

        # Get the list of potential winners
        winners = []
        for i in range(len(tanks)):
            if len( kills[tanks[0]] ) == len( kills[tanks[i]] ) and \
               tanks[0].isDead == tanks[i].isDead:
                winners.append(tanks[i])
            else:
                break
        winner = random.choice(winners)

        html = ['<html><body>',
                '<table><tr><th>Team<th>Kills<th>Cause of Death']
        for tank in tanks:
            if tank is winner:
                rowStyle = 'style="font-weight:bold; '\
                           'background-color:%s"' % tank._color
            else:
                rowStyle = 'style="background-color:%s"' % tank._color
            if name:
                name = xml.sax.saxutils.escape(tank.name)
            else:
                name = '#default'
            html.append('<tr %s><td>%s<td>%d<td>%s' % 
                        (rowStyle, 
                         name,
                         len(kills[tank]), 
                         xml.sax.saxutils.escape(tank.deathReason))) 

        html.append('</table><body></html>')

        # Write a blank file if the winner is a default tank..
        if winner.name != None:
            winnerFile.write(tanks[0].name)
        winnerFile.close()

        resultsFile.write('\n'.join(html))
        resultsFile.close()

    def _makeMovie(self):
        """Make the game movie."""

        movieCmd = ['ffmpeg', 
                    '-r', '10', # Set the framerate to 10/second
                    '-b', '8k', # Set the bitrate
                    '-i', '%s/%%05d.ppm' % self._imageDir, # The input files.
#                    '-vcodec', 'msmpeg4v2',
                    '%s/game.avi' % self._gameDir]

#        movieCmd = ['mencoder', 'mf://%s/*.png' % self._imageDir, 
#                    '-mf', 'fps=10', '-o', '%s/game.avi' % self._gameDir, 
#                    '-ovc', 'lavc', '-lavcopts', 
#                    'vcodec=msmpeg4v2:vbitrate=800']
        clearFrames = ['rm', '-rf', '%s' % self._imageDir]

        print 'Making Movie'
        subprocess.call(movieCmd)
#        subprocess.call(movieCmd, stderr=open('/dev/null', 'w'),
#                                  stdout=open('/dev/null', 'w'))
        subprocess.call(clearFrames)

    def _outputErrors(self, tank):
        """Output errors for each team."""
        if tank.name == None:
            return 
        
        if tank._program.errors:
            print tank.name, 'has errors'
            

        fileName = os.path.join(self._errorDir, quote(tank.name))
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
        self._imageDir = os.path.join(dir, 'frames')
        if not os.path.isdir(self._imageDir):
            os.mkdir( self._imageDir )
        self._playerDir = os.path.join(dir, 'ai', 'players')

    def _getDefaultAIs(self, dir, difficulty):
        """Load all the 'computer' controlled bot AIs for the given 
    difficulty."""
        defaultAIs = []

        path = os.path.join(dir, 'ai', difficulty)
        files = os.listdir( path )
        for file in files:
            if file.startswith('.'):
                continue

            path = os.path.join(dir, 'ai', difficulty, file)
            file = open( path ) 
            defaultAIs.append( file.read() )

        return defaultAIs
    
    def _getColors(self):
        """Get the team colors from the passwd file.  The passwd file location
    is set by self.TEAMS_FILE.  Returns a dictionary of players->color"""
        errorColor = '#ffffff'

        try:
            file = open(self.TEAMS_FILE)
        except:
            return {}.fromkeys(players, errorColor)

        colors = {}
        for line in file:
            try:
                team, passwd, color = map(unquote, line.split('\t'))
                colors[team] = '#%s' % color
            except:
                colors[team] = errorColor

        return colors

    def _getGameNum(self):
        """Figure out what game number this is from the past games played."""

        oldGames = os.listdir(self._resultsDir)
        games = []
        for dir in oldGames:
            try:
                games.append( int(dir) )
            except:
                continue

        games.sort(reverse=True)
        if games:
            return games[0] + 1
        else:
            return 0

if __name__ == '__main__':
    import sys, traceback
    try:
        p = Pflanzarr(sys.argv[1], sys.argv[2])
        p.run( int(sys.argv[3]) )
    except:
        traceback.print_exc()
        print "Usage: python2.6 Pflanzarr.py dataDirectory easy|medium|hard #turns"


