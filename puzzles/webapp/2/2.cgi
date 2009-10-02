#!/usr/bin/python

import cgi
import cgitb
cgitb.enable(context=10)

fields = cgi.FieldStorage()

print 'Content-Type: text/html'
print ''


print '''
<html>
	<head>
		<title>2</title>
		<link rel="stylesheet" type="text/css" href="ctf.css" media="all" />
	</head>
	<body>
		<div id="wrapper">
			<div id="content">
				<h1>Web Application Challenge 2</h1>
				<p>Through some manipulation or interpretation of this CGI script 
				and the HTML page(s) that it generates, a 10 character key can be 
				found.</p>
				<p><strong>Find the key!</strong></p>
				<p style="margin-top: 5em;">Question: How many geeks does it take to break a CGI?</p>
'''

# key = uq4G4dXrpx
if (fields.has_key('num')):
	print '''
				<p style="color: #fff;">You entered %d.</p>
	''' % int(fields['num'].value)

print '''
				<form method="get" action="two.py">
					Enter an integer: <input name="num" type="text" size="10" />
				</form>
			</div>
			<div id="footer">
				<p>Copyright &copy; 2009 LANS, LLC.</p>
			</div>
		</div>
	</body>
</html>
'''

