#! /usr/bin/lua

local koth = {}

-- cut -d$ANCHOR -f2- | grep -Fx "$NEEDLE"
function anchored_search(haystack, needle, anchor)
	for line in io.lines(haystack) do
		if (anchor) then
			pos = line:find(anchor)
			if (pos) then
				line = line:sub(pos+1)
			end
		end
		
		if (line == needle) then
			return true
		end
	end
	
	return false
end

function koth.anchored_search(haystack, needle, anchor)
	local ok, ret = pcall(anchored_search, haystack, needle, anchor)
	
	return ok and ret
end

function koth.page(title, body)
	print("Content-type: text/html")
	print()
	print("<!DOCTYPE html>")
	print("<html><head><title>" .. title .. "</title><link rel=\"stylesheet\" href=\"css/style.css\"></head>")
	print("<body><h1>" .. title .. "</h1>")
	if (body) then
		print("<section>")
		print(body)
		print("</section>")
	end
	print("</body></html>")
	os.exit(0)
end

--
-- We're going to rely on `bin/once` only processing files with the right number of lines.
--
function koth.award_points(team, category, points, comment)
	local filename = team .. "." .. category .. "." .. points
	local entry = team .. " " .. category .. " " .. points
	
	if (comment) then
		entry = entry .. " " .. comment
	end
	
	local ok = anchored_search("../state/points.log", entry, " ")
	if (not ok) then
		return false, "Points already awarded"
	end
	
	local f = io.open("../state/points.new/" .. filename, "a")
	if (not f) then
		return false, "Unable to write to points file"
	end
	
	f:write(os.time(), " ", entry, "\n")
	f:close()
	
	return true
end


return koth
