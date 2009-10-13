#!/usr/bin/python

import cgitb; cgitb.enable()
import os
from ctf import config

import Config

print(config.start_html('Tanks Results',
                        links_title='Tanks',
                        links=[('docs.cgi', 'Docs'),
                               ('results.cgi', 'Results'),
                               ('submit.html', 'Submit'),
                               ('errors.cgi', 'My Errors')]))
try:
    winner = open(os.path.join(Config.DATA_PATH, 'winner')).read()
except:
    winner = "No winner yet."

print "<H3>Last Winner: ", winner, '<H3>'
print "<H2>Results so far:</H2>"

try:
    games = os.listdir(os.path.join('results'))
except:
    print '<p>The results directory does not exist.'
    games = []

if not games:
    print "<p>No games have occurred yet."

gameNums = []
for game in games:
    try:
        gameNums.append( int(game) )
    except:
        continue

gameNums.sort(reverse=True)

# Don't include games that haven't completed
i = 0
num = str(gameNums[i])
for i in range(len(gameNums)):
    path = os.path.join( 'results', str(gameNums[i]), 'results.html') 
    if os.path.exists( path ):
        break
gameNums = gameNums[i:]

for num in gameNums:
    print '<p>%d - ' % num,
    print '<a href="results/%d/game.avi">v</a>' % num,
    print '<a href="results/%d/results.html">r</a>' % num

print(config.end_html())
