#!/usr/bin/python

import cgitb; cgitb.enable()
import os

import Config

print """Content-Type: text/html\n\n"""
print """<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">\n\n"""
head = open('head.html').read() % "Pflanzarr Results"
print head
print "<H1>Results</H1>"
print open('links.html').read()

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
        num = int(game)
        path = os.path.join( 'results', game, 'results.html')
        if os.path.exists( path ):
            gameNums.append( int(num) )
        else:
            continue

    except:
        continue

gameNums.sort(reverse=True)

for num in gameNums:
    print '<p>%d - ' % num,
    print '<a href="results/%d/game.avi">v</a>' % num,
    print '<a href="results/%d/results.html">r</a>' % num
