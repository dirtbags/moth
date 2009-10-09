#!/usr/bin/python

import cgitb; cgitb.enable()
import os
import sys
from ctf import config

try:
    from tanks import Program, setup, conditions, actions, docs
except:
    path = os.getcwd().split('/')
    path.pop()
    path.append('lib')
    sys.path.append(os.path.join('/', *path))
    import Program, setup, conditions, actions, docs

print(config.start_html('Tanks Documentation',
                        links_title='Tanks',
                        links=[('docs.cgi', 'Docs'),
                               ('results.cgi', 'Results'),
                               ('submit.html', 'Submit'),
                               ('errors.cgi', 'My Errors')]))

print Program.__doc__

print '<H3>Setup Actions:</H3>'
print 'These functions can be used to setup your tank.  Abuse of these functions has, in the past, resulted in mine sweeping duty.  With a broom.'
print "<P>"
docs.mkDocTable(setup.setup.values())

print '<H3>Conditions:</H3>'
print 'These functions are used to check the state of reality.  If reality stops being real, refer to chapter 5 in your girl scout handbook.<P>'
docs.mkDocTable(conditions.conditions.values())

print '<H3>Actions:</H3>'
print 'These actions are not for cowards.  Remember, if actions contradict, your tank will simply do the last thing it was told in a turn.  If ordered to hop on a plane to hell it will gladly do so.  If order to make tea shortly afterwards, it will serve it politely and with cookies instead.<P>'
docs.mkDocTable(actions.actions.values())

print(config.end_html())
