import subprocess
import os

class Jukebox:
   
    SALT = 'this is unreasonable.'
 
    def __init__(self, dataDir, tokens):

        self._dataDir = dataDir
        self.tokens = tokens

        self.station = None
        self._player = None

    def getStations(self):
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

    def play(self, user, token, station):
        """Switch to the given station, assuming it and the token are valid.
    raises a ValueError when either the station or token is unknown."""
        
        station = int(station)
        stations = self.getStations()
        if station not in stations:
            raise ValueError('Invalid Station (%s)' % station)

        if token not in self.tokens:
            raise ValueError('Invalid Token (%s)' % token)
        
        self.tokens.remove(token)
        self._changeStation( stations[station] )
            
    def mkToken(self, user):
        """Generate a token for the given user.  The token is a randomly 
    generate bit of text."""
        hash = sha256(self.SALT)
        hash.update(bytes(user, 'utf-8'))
        hash.update(bytes(str(time.time()), 'utf-8'))
        token = has.hex_digest()[:10]

        self.tokens.append(token)

        return token
 
    def _changeStation(self, file):
         
