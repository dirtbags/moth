#! /usr/bin/env python3

import cgitb; cgitb.enable()
import cgi
import fcntl
import string
from . import teams
from . import config

def head(title):
    return '''<?xml version="1.0" encoding="UTF-8" ?>
    <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
     "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
    <html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
      <head>
        <title>Team Registration</title>
        <link rel="stylesheet" href="%s" type="text/css" />
      </head>
      <body>
        <h1>%s</h1>
''' % (config.css, title)

def foot():
    return '''</body></html>'''

def main():
    print('Content-type: text/html')
    print()

    f = cgi.FieldStorage()

    team = f.getfirst('team', '')
    pw = f.getfirst('pw')
    confirm_pw = f.getfirst('confirm_pw')

    html = string.Template(config.start_html('Team Registration') +
                           ('''
        <p>
          Pick a short team name: you'll be typing it a lot.
        </p>

        <form method="post" action="%s">
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
        </form>''' % config.url('register.cgi')) +
                           config.end_html())

    if not (team and pw and confirm_pw):    # If we're starting from the beginning?
        html = html.substitute(team_error='',
                               pw_match_error='')
    elif teams.exists(team):
        html = html.substitute(team_error='Team team already taken',
                               pw_match_error='')
    elif pw != confirm_pw:
        html = html.substitute(team_error='',
                               pw_match_error='Passwords do not match')
    else:
        teams.add(team, pw)
        html = (config.start_html('Team registered') +
                ('<p>Congratulations, <samp>%s</samp> is now registered.  Go <a href="%s">back to the front page</a> and start playing!</p>' % (team, config.url(''))) +
                config.end_html())

    print(html)

if __name__ == '__main__':
    main()
