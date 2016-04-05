#! /usr/bin/env lua

package.path = "?.lua;cgi-bin/?.lua;www/cgi-bin/?.lua"

local koth = require "koth"

local max_by_cat = {}

local f = io.popen("ls " .. koth.path("packages"))
for cat in f:lines() do
	max_by_cat[cat] = 0
end
f:close()


for line in io.lines(koth.path("state/points.log")) do
	local ts, team, cat, points, comment = line:match("^(%d+) (%w+) ([%w-]+) (%d+) ?(.*)")
	points = tonumber(points) or 0
	
	-- Skip scores for removed categories
	if (max_by_cat[cat] ~= nil) then
		max_by_cat[cat] = math.max(max_by_cat[cat], points)
	end
end

local body = "<dl id=\"puzzles\">\n"
for cat, biggest in pairs(max_by_cat) do
	local points, dirname

	body = body .. "<dt>" .. cat .. "</dt>"
	body = body .. "<dd>"
	for line in io.lines(koth.path("packages/" .. cat .. "/map.txt")) do
		points, dirname = line:match("^(%d+) (.*)")
		points = tonumber(points)
		
		body = body .. "<a href=\"../" .. cat .. "/" .. dirname .. "/index.html\">" .. points .. "</a> "
		if (points > biggest) then
			break
		end
	end
	if (points == biggest) then
		body = body .. "<span title=\"Category Complete\">⁂</span>"
	end
	body = body .. "</dd>\n"
end
body = body .. "</dl>\n"
body = body .. "<fieldset><legend>Sandia Token:</legend>"
body = body .. "<p>Example: <samp>sandia:5:xylep-radar-nanox</samp></p>"
body = body .. "<form action='cgi-bin/token.cgi'>"
body = body .. "Team Hash: <input name='t'><br>"
body = body .. "Token: <input name='k'>"
body = body .. "<input type='submit'>"
body = body .. "</form>"
body = body .. "</fieldset>"
body = body .. "<p>Reloading this page periodically may yield updated puzzle lists.</p>"

koth.page("Open Puzzles", body)
