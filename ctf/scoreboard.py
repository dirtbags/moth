#!/usr/bin/env python3

import cgitb; cgitb.enable()
import os
import sys
from . import config
from . import teams
from . import points

flags_dir = config.get('global', 'flags_dir')
house_team = config.get('global', 'house_team')

def main():
    s = points.Storage()

    categories = [(cat, s.cat_points(cat)) for cat in s.categories()]

    print('Refresh: 10')
    print(config.start_html('Scoreboard', cls='wide'))
    print('<table class="scoreboard">')
    print('<tr>')
    print('<th>Overall</th>')
    for cat, score in categories:
        print('<th>')
        print('  %s (%d)' % (cat, score))
        try:
            fn = os.path.join(flags_dir, cat)
            team = open(fn).read() or house_team
            print('  <br/>')
            print('  <!-- flag: %s --> <span style="color: #%s" title="flag holder">%s</span>'
                  % (cat, teams.color(team), team))
        except IOError:
            pass
        print('</th>')
    print('</tr>')

    print('<tr>')
    print('<td><ol>')
    totals = []
    for team in s.teams:
        total = s.team_points(team)
        totals.append((total, team))
    for total, team in sorted(totals, reverse=True):
        print('<li><span style="color: #%s;">%s (%0.3f)</span></li>'
              % (teams.color(team), team, total))
    print('</ol></td>')
    for cat, total in categories:
        print('<td>')
        scores = sorted([(s.team_points_in_cat(cat, team), team) for team in s.teams])
        for score, team in scores:
            color = teams.color(team)
            print('<div style="height: %f%%; overflow: hidden; background: #%s; color: black;">' % (float(score * 100)/total, color))
            print('<!-- category: %s --> %s: %d' % (cat, team, score))
            print('</div>')
        print('</td>')
    print('</tr>')
    print('''</table>

        <p class="histogram">
          <img src="histogram.png" alt="scores over time" />
        </p>
''')
    print(config.end_html())

if __name__ == '__main__':
    main()

# Local Variables:
# mode: python
# End:
