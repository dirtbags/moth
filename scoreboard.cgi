#!/usr/bin/env python3

import cgitb; cgitb.enable()
import points

s = points.Storage('scores.dat')

teams = s.teams()
categories = [(cat, s.cat_points(cat)) for cat in s.categories()]
teamcolors = points.colors(teams)

print('Content-type: text/html')
print()

print('''<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
 "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
  <head>
    <title>CTF Scoreboard</title>
    <link rel="stylesheet" href="ctf.css" type="text/css" />
  </head>
  <body>
    <h1>Scoreboard</h1>
''')
print('<table>')
print('<tr>')
for cat, points in categories:
    print('<th>%s (%d)</th>' % (cat, points))
print('</tr>')

print('<tr>')
for cat, total in categories:
    print('<td style="height: 400px;">')
    scores = sorted([(s.team_points_in_cat(cat, team), team) for team in teams])
    for points, team in scores:
        color = teamcolors[team]
        print('<div style="height: %f%%; overflow: hidden; background: #%s; color: black;">' % (float(points * 100)/total, color))
        print('<!-- category: %s --> %s: %d' % (cat, team, points))
        print('</div>')
    print('</td>')
print('</tr>')
print('''</table>

    <img src="histogram.png" alt=""/>
  </body>
</html>''')

# Local Variables:
# mode: python
# End:
