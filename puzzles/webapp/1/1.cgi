#!/usr/bin/python

import cgi
import cgitb

print 'Content-Type: text/html'
print ''

print '''
<html>
	<head>
		<title>1</title>
<<<<<<< HEAD:puzzles/webapp/1/1.cgi
		<link rel="stylesheet" type="text/css" href="ctf.css" media="all" />
=======
		<link rel="stylesheet" type="text/css" href=",ctf.css" media="all" />
>>>>>>> 3c3c03775e5ac9a11abf581fbf8656e31d99ef42:puzzles/webapp/1/1.cgi
		<!-- key = ktFfb8R1Bw -->
	</head>
	<body>
		<div id="wrapper">
			<div id="content">
				<h1>Web Application Challenge 1</h1>
				<p>Through some manipulation or interpretation of this CGI script 
				and the HTML page(s) that it generates, a 10 character key can be 
				found.</p>
				<p><strong>Find the key!</strong></p>
			</div>
			<div id="footer">
				<p>Copyright &copy; 2009 LANS, LLC.</p>
			</div>
		</div>
	</body>
</html>
'''

