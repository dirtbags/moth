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
		<title>3</title>
		<link rel="stylesheet" type="text/css" href="../ctf.css" media="all" />
	</head>
	<body>
		<div id="wrapper">
			<div id="content">
				<h1>Web Application Challenge 3</h1>
				<p>Through some manipulation or interpretation of this CGI script 
				and the HTML page(s) that it generates, a 10 character key can be 
				found.</p>
				<p><strong>Find the key!</strong></p>

				<div class="vertsep"></div>
'''

PRODUCT_NAME = "Monkey of some kind"

def purchase_success(quantity):
	print '''
				<p>Congratulations, your order for %d "%s" has been placed.</p>
	''' % (quantity, PRODUCT_NAME)

# key = BRrHdtdADI
if fields.has_key('quantity') and fields.has_key('product') and fields['product'].value == PRODUCT_NAME:
	product  = fields['product'].value
	quantity = int(fields['quantity'].value)

	purchase_success(quantity)
else:
	print '''

				<h2>SALE: %s</h2>
				<p>Use the order form below to place an order.</p>

				<form method="get" action="3.py">
					How many would you like?
					<select name="quantity">
						<option value="12">12</option>
						<option value="24">24</option>
						<option value="48">48</option>
					</select>
					<br /><br />
					<input type="submit" value="Order!" />
					<input type="hidden" name="product" value="%s" />
				</form>
	''' % (PRODUCT_NAME, PRODUCT_NAME)

print '''

			</div>
			<div id="footer">
				<p>Copyright &copy; 2009 LANS, LLC.</p>
			</div>
		</div>
	</body>
</html>
'''

