#! /usr/bin/env lua

package.path = "?.lua;cgi-bin/?.lua;www/cgi-bin/?.lua"

local cgi = require "cgi"
local koth = require "koth"

local team = cgi.fields['t'] or ""
local category = cgi.fields['c'] or ""
local points = cgi.fields['p'] or ""
local answer = cgi.fields['a'] or ""

-- Defang category name; prevent directory traversal
category = category:gsub("[^A-Za-z0-9_]", "-")

-- Check answer
local needle = points .. " " .. answer
local haystack = koth.path("packages/" .. category .. "/answers.txt")
local found, err = koth.anchored_search(haystack, needle)

if (not found) then
	koth.page("Wrong answer", err)
end

local ok, err = koth.award_points(team, category, points)
if (not ok) then
	koth.page("Error awarding points",
	"<p>You got the right answer, but there was a problem trying to give you points:</p>" ..
	"<p>" .. err .. "</p>")
end

koth.page("Points awarded",
	"<p>" .. points .. " points for " .. team .. "!</p>" ..
	"<p><a href=\"../puzzles.html\">Back to puzzles</a></p>")
