#! /usr/bin/lua

function decode(str)
   local hexdec = function(h)
                     return string.char(tonumber(h, 16))
                  end
   str = string.gsub(str, "+", " ")
   return string.gsub(str, "%%(%x%x)", hexdec)
end

function decode_query(query)
   local ret = {}
   
   for key, val in string.gfind(query, "([^&=]+)=([^&=]+)") do
      ret[string.lower(decode(key))] = decode(val)
   end
       
   return ret
end

function escape(str)
   str = string.gsub(str, "&", "&amp;")
   str = string.gsub(str, "<", "&lt;")
   str = string.gsub(str, ">", "&gt;")
   return str
end

function djbhash(s)
   local hash = 5380
   for i=0,string.len(s) do
      local c = string.byte(string.sub(s, i, i+1))
      hash = math.mod(((hash * 32) + hash + c), 2147483647)
   end
   return string.format("%08x", hash)
end

function head(title)
   print("Content-type: text/html")
   print("")
   print("<!DOCTYPE html>")
   print("<html>")
   print("  <head>")
   print("    <title>")
   print(title)
   print("    </title")
   print('    <link rel="stylesheet" href="ctf.css" type="text/css">')
   print("  </head>")
   print("  <body>")
   print("    <h1>")
   print(title)
   print("    </h1>")
end

function foot()
   print("  </body>")
   print("</html>")
   os.exit()
end

if (os.getenv("REQUEST_METHOD") ~= "POST") then
   print("405 Method not allowed")
   print("Allow: POST")
   print("Content-type: text/html")
   print()
   print("<h1>Method not allowed</h1>")
   print("<p>I only speak POST.  Sorry.</p>")
end


inlen = tonumber(os.getenv("CONTENT_LENGTH"))
if (inlen > 200) then
   head("Bad team name")
   print("<p>That's a bit on the long side, don't you think?</p>")
   foot()
end
formdata = io.read(inlen)
f = decode_query(formdata)

team = f["t"]
if (not team) or (team == "dirtbags") then
   head("Bad team name")
   print("<p>Go back and try again.</p>")
   foot()
end
hash = djbhash(team)

if io.open(hash) then
   head("Team name taken")
   print("<p>Either someone's already using that team name,")
   print("or you found a hash collision.  Either way, you're")
   print("going to have to pick something else.</p>")
   foot()
end

f = io.open(hash, "w"):write(team)

head("Team registered")
print("<p>Team name: <samp>")
print(escape(team))
print("</samp></p>")
print("<p>Team token: <samp>")
print(hash)
print("</samp></p>")
print("<p><b>Save your team token somewhere</b>!")
print("You will need it to claim points.</p>")
foot()