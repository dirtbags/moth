#!/usr/bin/python

import os
import cgi
import cgitb
cgitb.enable(context=10)

#if os.environ.has_key('QUERY_STRING'):
#	os.environ['QUERY_STRING'] = ''

fields = cgi.FieldStorage()

import Cookie
c = Cookie.SimpleCookie(os.environ.get('HTTP_COOKIE', ''))

content = {
	'joke1' : '<p>An infinite number of mathematicians walk into a bar. The first one orders a beer. The second orders half a beer. The third, a quarter of a beer. The bartender says <em>You are all idiots!</em> and pours two beers.<p>',
	'joke2' : '<p>Two atoms are talking. One of them says <em>I think I lost an electron!</em> and the other says <em>Are you sure?</em> The first replies <em>Yeah, I am positive!</em></p>',
}

if c.has_key('content_name') and c.has_key('content'):
	k = c['content_name'].value
	try:
		c['content'] = content[k]
	except KeyError:
		c['content'] = '<p><em>key = s4nNlaMScV</em></p>'
else:
	c['content_name'] = 'joke1';
	c['content'] = content['joke1']


print 'Content-Type: text/html\n%s\n\n\n' % c
print ''

print '''
<html>
	<head>
		<title>7</title>
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

			function getContent() {
				content = readCookie("content");
				document.getElementById("stuff").innerHTML = content.substring(1, content.length-1);
			}

			window.onload = getContent;
		</script>
	</head>
	<body>
		<div id="wrapper">
			<div id="content">
				<h1>Web Application Challenge 7</h1>
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

