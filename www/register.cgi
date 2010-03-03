#! /usr/bin/python

import cgitb; cgitb.enable()
import cgi
import os
import fcntl
import string

from ctf import teams, html

def main():
    f = cgi.FieldStorage()

    team = f.getfirst('team', '')
    pw = f.getfirst('pw')
    confirm_pw = f.getfirst('confirm_pw')

    tmpl = string.Template('''
        <p>
          Pick a short team name: you'll be typing it a lot.
        </p>

        <form method="post" action="register.cgi">
          <fieldset>
            <legend>Registration information:</legend>

            <label>Team Name:</label>
            <input type="text" name="team" />
            <span class="error">$team_error</span><br />

            <label>Password:</label>
            <input type="password" name="pw" />
            <br />

            <label>Confirm Password:</label>
            <input type="password" name="confirm_pw" />
            <span class="error">$pw_match_error</span><br />

            <input type="submit" value="Register" />
          </fieldset>
        </form>''')

    if not (team and pw and confirm_pw):    # If we're starting from the beginning?
        body = tmpl.substitute(team_error='',
                               pw_match_error='')
    elif teams.exists(team):
        body = tmpl.substitute(team_error='Team team already taken',
                               pw_match_error='')
    elif pw != confirm_pw:
        body = tmpl.substitute(team_error='',
                               pw_match_error='Passwords do not match')
    else:
        teams.add(team, pw)
        body = ('<p>Congratulations, <samp>%s</samp> is now registered.  Go <a href="/">back to the front page</a> and start playing!</p>' % cgi.escape(team))

    html.serve('Team Registration', body)

if __name__ == '__main__':
    import sys, codecs

    sys.stdout = codecs.getwriter('utf-8')(sys.stdout)

    main()
