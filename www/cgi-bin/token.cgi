#! /usr/bin/env lua

package.path = "?.lua;cgi-bin/?.lua;www/cgi-bin/?.lua"

local cgi = require "cgi"
local moth = require "moth"

local team = cgi.fields['t'] or ""
local token = cgi.fields['k'] or ""

-- Check answer
local needle = token
local haystack = moth.path("state/tokens.txt")
local found, err = moth.anchored_search(haystack, needle)

if (not found) then
	moth.page("Unrecognized token", err)
end

local category, points = token:match("^(.*):(.*):")
if ((category == nil) or (points == nil)) then
	moth.page("Unrecognized token", "Something doesn't look right about that token")
end
points = tonumber(points)

-- Defang category name; prevent directory traversal
category = category:gsub("[^A-Za-z0-9]", "-")

local ok, err = moth.award_points(team, category, points, token)
if (not ok) then
	moth.page("Error awarding points",
	"<p>You entered a valid token, but there was a problem trying to give you points:</p>" ..
	"<p>" .. err .. "</p>")
end

moth.page("Points awarded",
	"<p>" .. points .. " points for " .. team .. "!</p>" ..
	"<p><a href=\"../puzzles.html\">Back to puzzles</a></p>")
