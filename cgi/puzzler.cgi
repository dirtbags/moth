#! /usr/bin/lua

local cgi = require "cgi"

cgi.init()
fields = cgi.fields()
print("Content-type: text/html")
print()
print("<pre>")
print(fields["t"])
print("</pre>")
