#! /usr/bin/env lua

package.path = "?.lua;cgi-bin/?.lua;www/cgi-bin/?.lua"


local cgi = require "cgi"
local moth = require "moth"

local team = cgi.fields["n"] or ""
local hash = cgi.fields["h"] or ""

hash = hash:match("[0-9a-f]*")

if ((hash == "") or (team == "")) then
	moth.page("Invalid Entry", "Oops! Are you sure you got that right?")
elseif (not moth.anchored_search(moth.path("state/assigned.txt"), hash)) then
	moth.page("Invalid Hash", "Oops! I don't have a record of that hash. Did you maybe use capital letters accidentally?")
end

local f = io.open(moth.path("state/teams/" .. hash))
if (f) then
	f:close()
	moth.page("Already Exists", "Your team has already been named! Maybe somebody on your team beat you to it.")
end

local f, err = io.open(moth.path("state/teams/" .. hash), "w+")
if (not f) then
	moth.page("Kersplode", err)
end
f:write(team)
f:close()

moth.page("Success", "Okay, your team has been named and you may begin using your hash!")
