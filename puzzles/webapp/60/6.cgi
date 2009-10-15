#!/usr/bin/python

import os
import cgi
import cgitb
cgitb.enable(context=10)

#if os.environ.has_key('QUERY_STRING'):
#	os.environ['QUERY_STRING'] = ''

fields = cgi.FieldStorage()

import Cookie
c = Cookie.SimpleCookie()
c['key'] = 'QJebByJaKX'
c['content'] = '<p><em>Maybe I should have used sessions...</em></p>'

print 'Content-Type: text/html\n%s\n\n\n' % c
print ''

print '''
<html>
	<head>
		<title>6</title>
		<link rel="stylesheet" type="text/css" href=",ctf.css" media="all" />
		<script type="text/javascript">
			function readCookie(key) {
				var s = key + '=';
				var toks = document.cookie.split(';');
				for (var i = 0; i < toks.length; i++) {
					var tok = toks[i];
					while (tok.charAt(0) == ' ') {
						tok = tok.substring(1, tok.length);
					}
					if (tok.indexOf(s) == 0) {
						return tok.substring(s.length, tok.length);
					}
				}
				return null;
			}

			function setContent() {
				content = readCookie("content");
				document.getElementById("stuff").innerHTML = content.substring(1, content.length-1);
			}

			window.onload = setContent;
		</script>
	</head>
	<body>
		<div id="wrapper">
			<div id="content">
				<h1>Web Application Challenge 6</h1>
				<p>Through some manipulation or interpretation of this CGI script 
				and the HTML page(s) that it generates, a 10 character key can be 
				found.</p>
				<p><strong>Find the key!</strong></p>

				<div class="vertsep"></div>
				<div id="stuff"></div>
'''

print '''
			</div>
			<div id="footer">
				<p>Copyright &copy; 2009 LANS, LLC.</p>
			</div>
		</div>
	</body>
</html>
'''

