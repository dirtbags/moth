#!/usr/bin/env python

import cgi
import cgitb
import os
import time

cgitb.enable()

form = cgi.FieldStorage()
client = os.environ["REMOTE_ADDR"]

fields = {
	'affiliation' : ['nnsa', 'doe', 'dod', 'otherfed', 'state', 'private', 'other'],
	'hostforensics' : ['has', 'doesnt_have_can_get', 'doesnt_have_cant_get'],
	'netforensics' : ['has', 'doesnt_have_can_get', 'doesnt_have_cant_get'],
	'reversing' : ['has', 'doesnt_have_can_get', 'doesnt_have_cant_get'],
	'regularcollab' : ['0', '1', '2', '3', '4', '5+'],
	'collab' : ['0', '1', '2', '3', '4', '5+'],
	'incident' : ['0', '1', '2', '3', '4', '5+'],
	'channels' : ['official', 'unofficial'],
	'helpfulone' : ['tracer', 'cons', 'vtc', 'tc', 'irc'],
	'helpfultwo' : ['tracer', 'cons', 'vtc', 'tc', 'irc'],
	'helpfulthree' : ['tracer', 'cons', 'vtc', 'tc', 'irc'],
	'helpfulfour' : ['tracer', 'cons', 'vtc', 'tc', 'irc'],
	'helpfulfive' : ['tracer', 'cons', 'vtc', 'tc', 'irc'],
	'toolset' : ['0', '1', '2', '3', '4'],
	'overall' : ['0', '1', '2', '3', '4'],
	'comments' : []
	}

def validate(form):

	for k,v in fields.items():
		if len(v) and form.getfirst(k) not in v:
			return False

	vals = []
	for k in ['helpfulone', 'helpfultwo', 'helpfulthree', 'helpfulfour', 'helpfulfive']:
		if form.getfirst(k) in vals:
			return False
		vals.append(form.getfirst(k))

	return True

print 'Content-Type: text/html'
print ''

print '''
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en-GB">
<head>
	<title>CyberTracer Collaboration Survey</title>
	<meta http-equiv="Content-Type" content="application/xhtml+xml; charset=utf-8" />
	<link rel="stylesheet" href="survey.css" type="text/css" />
</head>
<body>
	<div id="wrapper">

		<div id="header">
			<h1>Cyber Security Collaboration Survey &mdash; Tracer FIRE II</h1>
		</div>

		<div id="content">
'''

if validate(form):
	results = [client, str(time.time())]

	for k in fields.keys():
		val = form.getfirst(k) or ''
		if k == 'comments':
			val = val.replace(',', ' ')
			val = val.replace(':', ' ')
			val = val.replace('\n', ' ')
			val = val.replace('\r', ' ')
		results.append('%s:%s' % (k, val))

	f = open('/var/lib/ctf/survey/%s' % client, 'a')
	f.write(','.join(results) + '\n')
	f.close()

	print '<p><b>SUCCESS!</b> Your survey submission has been accepted. Please <b>do not</b> retake the survey. Thanks!</p>'
else:
	print '''
		<p><b>FAIL!</b> It looks like you bypassed the client-side validation of the survey! That's too easy and the contest
		hasn't even begun yet! Would you please go back and just take the survey? It is very important!</p>
	'''

print '''
		</div>
	</div>
</body>
</html>
'''
