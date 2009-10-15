#!/usr/bin/python

import os
import cgi
import cgitb
import sqlite3
cgitb.enable(context=10)

if os.environ.has_key('QUERY_STRING'):
	os.environ['QUERY_STRING'] = ''

fields = cgi.FieldStorage()

q = None
if fields.has_key('q'):
	q = fields['q'].value

if q is not None:
	print 'Content-Type: text/html\n'
	try:
		db = sqlite3.connect(',zomg.sqlite3')
		cur = db.cursor()
		cur.execute(q)
		results = cur.fetchall()
		
		print '<table>'
		for r in results:
			print '<tr>'
			for thing in r:
				print '<td>%s</td>' % thing
			print '</tr>'
		print '</table>'
			
	except Exception:
		print '<p class="error">Invalid query: %s</p>' % q

else:
	print 'Content-Type: text/html\n'
	print ''

	print '''
	<html>
		<head>
			<title>8</title>
			<link rel="stylesheet" type="text/css" href=",ctf.css" media="all" />
			<script type="text/javascript">
	
				function buildQuery(table_name, result_limit) {
					var q = "SELECT * FROM " + table_name + " LIMIT " + result_limit;
					return q;
				}
	
				function getXHRObject() {
					var xhr = null;
					try {
						xhr = new XMLHttpRequest();
					}
					catch (ex) {
						try {
							xhr = new ActiveXObject("msxml2.XMLHTTP");
						}
						catch (ex) {
							alert("Browser does not support AJAX!")
							return null;
						}
					}
					return xhr;
				}

				function sendXHRPost(xhr, url, params) {
					xhr.open("POST", url, true);
					xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
					xhr.setRequestHeader("Content-length", params.length);
					xhr.setRequestHeader("Connection", "close");
					xhr.send(params)
				}
	
				function doQuery(q) {
					var xhr = getXHRObject();
					if (xhr != null) {
						var url = "8.cgi";
						var params = "q=" + q;
						xhr.onreadystatechange = function() {
							if (xhr.readyState == 4) {
								var response = xhr.responseText;
								var d = document.getElementById("results");
								d.innerHTML = response;
							}
						}
						sendXHRPost(xhr, url, params);
					}
				}

				function submitForm() {
					var f = document.getElementById("the_form");
					var table_name = f.tname.value;
					var result_limit = f.rlimit.value;

					var q = buildQuery(table_name, result_limit);

					doQuery(q);

					return false;
				}

			</script>
		</head>
		<body>
			<div id="wrapper">
				<div id="content">
					<h1>Web Application Challenge 8</h1>
					<p>Through some manipulation or interpretation of this CGI script 
					and the HTML page(s) that it generates, a 10 character key can be 
					found.</p>
					<p><strong>Find the key!</strong></p>

					<div class="vertsep"></div>
					<h2>Database Query Wizard</h2>
					<p>Use the form below to retrieve data from the database. Select the
					type of data that you would like to view and the number of database
					entries to retrieve and then click on the &quot;Query&quot; button.</p>
	
					<form id="the_form" action="" method="POST" onsubmit="return submitForm()">
						<br />
						Topic: <select name="tname">
							<option value="jokes">Jokes</option>
							<option value="pics">Pictures</option>
						</select>
						<br /><br />
						# Results: <select name="rlimit">
							<option value="1">1</option>
							<option value="2">2</option>
							<option value="3">3</option>
							<option value="4">4</option>
							<option value="5">5</option>
						</select>
						<br /><br />
						<input type="submit" value="Query" />
					</form>

					<div id="results"></div>
				</div>
				<div id="footer">
					<p>Copyright &copy; 2009 LANS, LLC.</p>
				</div>
			</div>
		</body>
	</html>
	'''

