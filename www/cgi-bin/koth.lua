#! /usr/bin/env lua

local koth = {}

-- cut -d$ANCHOR -f2- | grep -Fx "$NEEDLE"
function koth.anchored_search(haystack, needle, anchor)
	local f, err = io.open(haystack)
	if (not f) then
		return false, err
	end
	
	for line in f:lines() do
		if (anchor) then
			pos = line:find(anchor)
			if (pos) then
				line = line:sub(pos+1)
			end
		end
		
		if (line == needle) then
			f:close()
			return true
		end
	end
	
	f:close()
	return false
end

function koth.page(title, body)
	if (os.getenv("REQUEST_METHOD")) then
		print("Content-type: text/html")
		print()
	end
	print("<!DOCTYPE html>")
	print("<html><head><title>" .. title .. "</title><link rel=\"stylesheet\" href=\"../style.css\"><meta name=\"viewport\" content=\"width=device-width\"></head>")
	print("<body><h1>" .. title .. "</h1>")
	if (body) then
		print("<section>")
		print(body)
		print("</section>")
	end
	print('<section id="sponsors">')
	print('<img src="../images/lanl.png" alt="Los Alamos National Laboratory">')
	print('<img src="../images/doe.png" alt="US Department Of Energy">')
	print('<img src="../images/sandia.png" alt="Sandia National Laboratories">')
	print('</section>')
	print("</body></html>")
	os.exit(0)
end

--
-- We're going to rely on `bin/once` only processing files with the right number of lines.
--
function koth.award_points(team, category, points, comment)
	team = team:gsub("[^0-9a-f]", "-")
	if (team == "") then
		team = "-"
	end

	local filename = team .. "." .. category .. "." .. points
	local entry = team .. " " .. category .. " " .. points
	
	if (comment) then
		entry = entry .. " " .. comment
	end
	
	local f = io.open(koth.path("state/teams/" .. team))
	if (f) then
		f:close()
	else
		return false, "No such team"
	end
	
	local ok = koth.anchored_search(koth.path("state/points.log"), entry, " ")
	if (ok) then
		return false, "Points already awarded"
	end
	
	local f = io.open(koth.path("state/points.new/" .. filename), "a")
	if (not f) then
		return false, "Unable to write to points file"
	end
	
	f:write(os.time(), " ", entry, "\n")
	f:close()
	
	return true
end

-- Most web servers cd to the directory containing the CGI.
-- Not uhttpd.

koth.base = ""
function koth.path(p)
	return koth.base .. p
end

-- Traverse up to find assigned.txt
for i = 0, 5 do
	local f = io.open(koth.path("state/assigned.txt"))
	if (f) then
		f:close()
		break
	end
	koth.base = koth.base .. "../"
end

return koth
