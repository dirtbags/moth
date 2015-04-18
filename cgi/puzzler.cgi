#! /usr/bin/lua

local cgi = require "cgi"


-- Read in team and answer



print("Content-type: text/html")
print()
print("<pre>")
print(cgi.fields["t"])
print("</pre>")
