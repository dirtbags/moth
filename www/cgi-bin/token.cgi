#! /usr/bin/env lua

package.path = "?.lua;cgi-bin/?.lua;www/cgi-bin/?.lua"

local cgi = require "cgi"
local koth = require "koth"

local team = cgi.fields['t'] or ""
local token = cgi.fields['k'] or ""

-- Defang category name; prevent directory traversal
category = category:gsub("[^A-Za-z0-9]", "-")

-- Check answer
local needle = token
local haystack = koth.path("tokens.txt")
local found, err = koth.anchored_search(haystack, needle)

if (not found) then
	koth.page("Unrecognized token", err)
end

local category, points = token.match("^(.*):(.*):")
if ((category == nil) || (points == nil)) then
	koth.page("Unrecognized token", "Something doesn't look right about that token")
end
points = tonumber(points)

local ok, err = koth.award_points(team, category, points, token)
if (not ok) then
	koth.page("Error awarding points",
	"<p>You entered a valid token, but there was a problem trying to give you points:</p>" ..
	"<p>" .. err .. "</p>")
end

koth.page("Points awarded",
	"<p>" .. points .. " points for " .. team .. "!</p>" ..
	"<p><a href=\"../puzzles.html\">Back to puzzles</a></p>")
