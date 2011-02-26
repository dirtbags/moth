#! /usr/bin/lua

require("lfs")

BASEDIR = "/var/tmp/wopr"
POST_MAX = 512

method = os.getenv("REQUEST_METHOD")
if (method == "POST") then
   local CL = tonumber(os.getenv("CONTENT_LENGTH")) or 0
   if (CL > POST_MAX) then
      CL = POST_MAX
   end
   function getc()
      if (CL > 0) then
         CL = CL - 1
         return io.read(1)
      else
         return nil
      end
   end
elseif (method == "GET") then
   local query = os.getenv("QUERY_STRING") or ""
   local query_pos = 0
   local query_len = string.len(query)
   if (query_len > POST_MAX) then
      query_len = POST_MAX
   end
   function getc()
      if (query_pos < query_len) then
         query_pos = query_pos + 1
         return string.sub(query, query_pos, query_pos)
      else
         return nil
      end
   end
else
   print("405 Method not allowed")
   print("Allow: GET POST")
   print("Content-Type: text/plain")
   print()
   print("I only do GET and POST.")
   os.exit(0)
end

function read_hex()
   local a = getc() or 0
   local b = getc() or 0

   return string.char(tonumber(a, 16)*16 + tonumber(b, 16))
end

function cgi_item()
   local val = ""

   while (true) do
      local c = getc()
      if ((c == nil) or (c == "=") or (c == "&")) then
         return val
      elseif (c == "%") then
         c = read_hex()
      elseif (c == "+") then
         c = " "
      end
      val = val .. c
   end
end

function escape(s)
   s = string.gsub(s, "&", "&amp;")
   s = string.gsub(s, "<", "&lt;")
   s = string.gsub(s, ">", "&gt;")
   return s
end

f = {}
while (true) do
   local key = cgi_item()
   local val = cgi_item()

   if (key == "") then
      break
   end
   f[key] = val
end



-- lua doesn't seed its PRNG and provides nothing other than
-- time in seconds.  If you're on Windows, go fish.
do
   local seed = 0
   r = io.open("/dev/urandom") or io.open("/dev/random")
   for i = 1, 4 do
      seed = seed*256 + string.byte(r:read(1))
   end
   r:close()
   math.randomseed(seed)
end

-- Get or create Session ID
sid = f["s"] or ""
if (sid == "") then
   sid = string.format("%08x.%04x", os.time(), math.random(65535))
end
dirname = BASEDIR .. "/" .. sid

-- Send back a page
function reply(text, prompt, ...)
   print("Content-type: text/xml")
   print()
   print("<document>")
   print("  <sessionid>" .. sid .. "</sessionid>")
   print("  <response>" .. escape(text or "") .. "</response>")
   print("  <prompt>" .. escape(prompt or ">") .. "</prompt>")
   if (arg[1]) then
      print("  <error>" .. escape(arg[1]) .. "</error>")
   end
   print("</document>")
   os.exit(0)
end




--
-- Database functions
--

function get(key, ...)
   local fn = string.format("%s/%s", dirname, key)
   local f = io.open(fn)
   if (not f) then
      return arg[1]
   else
      local ret = f:read(4000) or ""
      f:close()
      return ret
   end
end

function set(key, ...)
   local fn = string.format("%s/%s", dirname, key)
   local f

   -- Lazy mkdir to save a few inodes
   lfs.mkdir(dirname)

   f = io.open(fn, "w")
   if not f then
      error("Unable to write " .. fn)
   end
   f:write(arg[1] or "")
   f:close()
end

function del(key)
   local fn = string.format("%s/%s", dirname, key)
   os.remove(fn)
end


--
-- A string splitter
--
function string:split(...)
   local sep = arg[1] or " "
   local ret = {}
   local start = 1

   while true do
      local first, last = self:find(sep, start)
      if not first then
         break
      end
      table.insert(ret, self:sub(start, first - 1))
      start = last + 1
   end
   table.insert(ret, self:sub(start))

   return ret
end


-------------------------------------
--
-- WOPR-specific stuff
--

-- A list of all hosts, by name
hosts_by_name = {}


Host = {}

function Host:new(name, ...)
   local o = {}
   setmetatable(o, self)
   self.__index = self
   o.name = name
   o.prompt = (arg[1] or ">")
   o.obuf = {}
   o.history = {}

   hosts_by_name[name] = o
   return o
end

function Host:add_commands(t)
   local cmds = {}
   local k, v

   for k,v in pairs(self.commands) do
      cmds[k] = v
   end
   for k,v in pairs(t) do
      cmds[k] = v
   end
   self.commands = cmds
end

function Host:get(key, ...)
   return get(self.name .. "." .. key, arg[1])
end

function Host:set(key, ...)
   return set(self.name .. "." .. key, arg[1])
end

function Host:del(key)
   return del(self.name .. "." .. key)
end

function Host:writeln(...)
   table.insert(self.obuf, (arg[1] or ""))
end

function Host:login(...)
   set("host", self.name)
   reply(arg[1] or self.motd, self.prompt)
end

function Host:cmd_help()
   local k, v
   self:writeln("Available commands:")
   for k,v in pairs(self.commands) do
      if (v[1]) then
         local s = string.format("%-15s  %s", k, v[1])
         self:writeln(s)
      end
   end
end

function Host:cmd_history()
   local k, v
   for k,v in ipairs(self.history) do
      self:writeln(string.format("%5d  %s", k, v))
   end
end

-- Call self:handle(req) and return what to send back
function Host:handle_request(req)
   local t = ""
   local k, v

   self:handle(req)

   for k,v in ipairs(self.obuf) do
      t = t .. v .. "\n";
   end
   reply(t, self.prompt)
end

-- Handle a request
function Host:handle(req)
   self:do_cmd(req)
end

-- Run a command or return an error
function Host:do_cmd(req)
   local argv = req:split()
   local cmd = self.commands[argv[1]:lower()]

   if (argv[1] == "") then
      return
   end

   -- First, update history
   if self.history then
      local h = self:get("history")
      if h then
         self.history = h:split("\n")
      end
      table.insert(self.history, req)
      self:set("history", table.concat(self.history, "\n"))
   end

   -- Now run the command
   if cmd then
      if cmd[2] then
         cmd[2](self, argv)
      else
         self:writeln("ERROR: no function defined")
      end
   else
      self:writeln("Unknown command")
   end
end

-- List of commands, with help string (nil hides from help)
Host.commands = {
   ["?"] = {nil, Host.cmd_help},
   ["help"] = {"List available commands", Host.cmd_help},
   ["history"] = {"Display command history", Host.cmd_history},
}


--
-- Login screen
--
Login = Host:new("login", "Enter 12-digit access code:")
Login.motd = [[

┃┃┃┏━┃┏━┃┏━┃
┃┃┃┃ ┃┏━┛┏┏┛
━━┛━━┛┛  ┛ ┛  3.0
War Operations Plan Response
New Khavistan Ministry of Ministries

This computer system is the property of the government of New Khavistan.
It is for authorized use only.  Any or all uses of this system and all
files on this system are monitored and logged.  By using this system,
the user consents to such monitoring. Users who do not consent to such
monitoring will be dispatched with the New Khavistan fiber-optic bullet
delivery system.

Users should have no expectation of privacy as to any communication on
or information stored within the system, including but not limited to
information stored within your brain, DNA, government tooth implants,
or tinfoil hat.

Unauthorized or improper use of this system may result in gnomes pooping
in your underpants.  By continuing to use this system, you indicate your
awareness of and consent to these terms and conditions of use.  LOG OFF
IMMEDIATELY if you do not agree to the conditions stated in this
warning.

]]

function Login:handle(req)
   if (string.len(req) > 20) then
      -- Log them in to wopr
      Wopr:login([[
FLAGRANT SYSTEM ERROR: Memory segmentation violation
Returning to command subsystem [wopr:xipir-cavud-libux]
]])
   else
      if (req == "joshua") then
         self:writeln("wopr:xirak-zoses-gefox")
      elseif (req ~= "") then
         self:writeln("Incorrect code")
      end
   end
end

function Login:login(...)
   -- Since login is the default, we can *unset* host.
   -- This has the nice property of not allocating any
   -- storage for people who never make it past the front door.
   del("host")
   reply(arg[1] or self.motd, self.prompt)
end


--
-- Bulletin Board (bb) subsystem
--
Bb = Host:new("bb", "[N]ext, (P)rev, (Q)uit, msg#:")
Bb.posts = {[[
WOPR operational!  =====  administrator  =====  Aug 16 2003

Welcome to WOPR system.  Authorized by FLD-853, system will be linked
with all critical New Khavistan technical infrastructure.  Mandated by
FLD 897 will be full compliance by 2007.  Waiting time, following
services may be used: telecommunications, traffic control devices,
payroll, and strategic missile offensive control.

Finding any problems with this system, simply fill and submit form
CPW-190.  Royal Ministry Of Technology processes all properly-reported
issues with utmost haste and concern.

::: FLD-711 Restricted Distribution :::
]], 
[[
Clock problem  =====  administrator  =====  Sep 35, 1568

Royal Ministry Of Technology is aware of recent problems with system
time.  We strive to rectify this problem.  Thank you for patience with
issue.

::: FLD-711 Restricted Distribution :::
]], 
[[
System Overhaul  =====  Krdznyklyk  =====  Dec 12, 2003

The entire system is being overhauled to fix security holes exposed by a
recent attack on our systems. Your presence is requested at an all-hands
briefing to roll out the new system this afternoon at 1500 hrs.
]], 
[[
Drill  =====  wopr:xofic-belid-civox  =====  Dec 15, 2003

There will be a drill to test our combat readiness today. Make sure you
are familiar with the proper procedures to complete the task. Follow
proper drilling procedures.
]], 
[[
Passwords  =====  administrator  =====  Jan 12, 2004

All passwords have been modified to end with character "!", to bring
WOPR compliant with Fearless Leader Directive 1138 "standards for secure
passwords".  For example, a password once "cascade", is becoming
"cascade!".

Recent change improves important system resilience against attack from
enemies.  Your gracious understanding and support of New Khavistan is
appreciated.

::: FLD-711 Restricted Distribution :::
]], 
[[
ICBM control  =====  administrator  ====  Feb 2, 2004

Because of new security protocols in FLD-1205, you must now type,
"override on" to get access to ICBM commands.

::: FLD-711 Restricted Distribution :::
]], 
[[
attcon command  =====  administrator  =====  Feb 14, 2004

The WOPR system has been updated to include a new command!

Purpose: To make it possible to change the attack condition level for
all troops in the New Khavistan republic, simple and unified.

Subsystems Affected: Small subset of WOPR subsystems

Usage: Type "attcon" then number pertaining to correct readiness level
(1-5). Also can get attcon level by only typing attcon.

Expected loss of service: 30 min while WOPR system recycles

::: FLD-711 Restricted Distribution :::
]], 
[[
Syistem Adimn  =====  ACTION REQUIRED  ====  0573, Dec 32 2004

ACTION REQUIRED

Dear empyoyee,

We have an important message for you form your commanding
officer. Please click here to veiw the messsage.
]], 
[[
Alert: message attack  =====  administrator  =====  Jun 20, 2004

NOT TO CLICK ON LINK.

Perverted computer attackers from libellous Republic of Dweezil break
WOPR system security and try to subvert glorious nation with tricksy
electronic link.  NT TO CLICK.

New Khavistan secret infantry being dispatched to deal with computer
threat from libellous Republic of Dweezil.
]], 
[[
Security Breach  =====  administrator  =====  Jun 21, 2004

Who click link?  Now is time to come forward and accept judicious
punishment from Fearless Leader.
]], 
[[
Mandatory Training  =====  administrator  =====  Jun 24, 2004

Phishing awareness training today at 1500 hrs
]], 
[[
FLD-1327  =====  administrator  =====  Aug 10, 2004

FLD-1327 extends target date for full WOPR integration until June 22,
2058.  Meantime WOPR subsystem continue to operate.

::: FLD-711 Restricted Distribution :::
]],

-- Message ID #-5: a snippet of the WOPR command program
[-1] = "^A^@^@<8B>U<D4><E8>wopr:xetil-rokak-robyx<AD><FF><FF><FF>^O<B6>MЉËEօ<C0>uG<84>",
[-2] = "<C3>^A<90><8D>t&^@<E9>@<FF><FF><FF><8D>v^@<83>",
[-3] = "[^_]Ð<8D>t&^@<89>^\$<E8>Ѐ^@^@<8B>=<A4><C2>^E^H<80>;/<89>",
[-4] = "<85><D2>^?׃<C4>^T1<C0>[]Ít&^@<B8>^A^@^@^@븋C^D<89>",
[-5] = [[
^@^@^@^@^@^@on", n] => set attcon = n
                 msg "attcon set to" n
["attcon enid"] => set_launch_trigger(1)
["attcon dennis"] => set_launch_trigger(0)
["bb"] => call_subsys bb
["test"] => msg "test out<?tE<t<C3><C7>^D$<9F><9B>^E^H<E8>b<84>A<[,[^_]WVS1
]],
[-6] = "^@^@1^E^@^@2^E^@^@3^E^@^@4",
[-7] = "^^F^@^@!^F^@^@$^F^@^@&^F^@^@)",
[-8] = "^@",
}

Bb.motd = [[
                WOPR Message Board
====================================================
   [N]ext message
   (P)revious message
   (Q)uit
   Enter message number to jump to that message
]]

function Bb:read(inc)
   local msgid = tonumber(self:get("msgid")) or 0
   msgid = msgid + inc
   self:jump(msgid)
end

function Bb:jump(msgid)
   self:set("msgid", msgid)

   self:writeln("::::::::::::::::::::::::: Message #" .. tostring(msgid))
   self:writeln()
   self:writeln(self.posts[msgid])
end

function Bb:do_cmd(req)
   local n = tonumber(req)
   if (req == "") then
      self:cmd_next()
   elseif n then
      self:jump(n)
   else
      Host.do_cmd(self, req)
   end
end

function Bb:cmd_next(argv)
   self:read(1)
end

function Bb:cmd_prev(argv)
   self:read(-1)
end

function Bb:cmd_help(argv)
   self:writeln(self.motd)
end

function Bb:cmd_quit(argv)
   Wopr:login()
end

Bb.commands = {
   ["?"] = {nil, Bb.cmd_help},
   ["n"] = {nil, Bb.cmd_next},
   ["p"] = {nil, Bb.cmd_prev},
   ["q"] = {nil, Bb.cmd_quit},
}

--
-- The WOPR host
--
Wopr = Host:new("wopr", "WOPR%")
Wopr.history = {
   'subsys comm',
   'exit',
   'subsys comm',
   'bb',
   'subsys comm',
   'exit',
   'bb',
   'subsys comm',
   'exit',
   'hlep',
   'help',
   'bb',
   'help',
   'subsys comm',
   'exit',
}

Wopr.motd = ""

function Wopr:cmd_subsys(argv)
   local sys = argv[2]

   if not sys then
      self:writeln("Usage: subsys SYSTEM")
   elseif sys == "?" then
      local k, v
      for k,v in pairs(hosts_by_name) do
         self:writeln(k)
      end
   else
      h = hosts_by_name[sys] 
      if not h then
         self:writeln("No such subsystem (? to list)")
      else
         h:login()
      end
   end
end

function Wopr:cmd_bb(argv)
   Bb:login()
end

function Wopr:attcon()
   return tonumber(self:get("attcon") or 5)
end

-- This command should feel really shoddy: it was written
-- in-house by the New Khavistan Ministry of Technology.
function Wopr:cmd_attcon(argv)
   if argv[2] == "enid" then
      self:writeln("[[[ LAUNCH TRIGGER ENABLED ]]]")
      self:writeln("wopr:xelev-lepur-pozyx")
      self:set("launch")
   elseif argv[2] == "dennis" then
      self:writeln("[[[ LAUNCH TRIGGER DISABLED ]]]")
      self:del("launch")
   elseif argv[2] then
      local v = tonumber(argv[2]) or 5
      self:set("attcon", v)
      self:writeln("attcon set to " .. tostring(v))
   else
      self:writeln(tostring(self:attcon()))
   end
end

-- Some test code they didn't remove
function Wopr:cmd_test(argv)
   self:writeln("test output:")
   self:writeln("  EIGEN58")
   self:writeln("  sub_malarkey reached")
   self:writeln("  DEBUG:453:wopr:xocom-bysik-mapix")
   self:writeln("$$END")
end

Wopr:add_commands{
   ["subsys"] = {"Connect to subsystem", Wopr.cmd_subsys},
   ["bb"]     = {"Read bulletin board", Wopr.cmd_bb},
   ["attcon"] = {"[Place command description here]", Wopr.cmd_attcon},
   ["test"]   = {nil, Wopr.cmd_test},
}

--hosts["wopr"] = Wopr

--
-- Communications subsystem
--
Comm = Host:new("comm", "COMSYS>")

Comm.motd = [[
_____IMPORTANT_____

IBM 3750 used for main switchboard in captiol building is currently
running at half capacity until relays arrive.  Please to remember not
patching trunks to switch!
]]

function Comm:cmd_exit(argv)
   Wopr:login()
end

function Comm:cmd_status(argv)
   self:writeln("[Not yet implemented]")
   self:writeln("wopr:xoroc-hunaz-vyhux")
end

Comm:add_commands{
   ["status"] = {"Display phone system status", Comm.cmd_status},
   ["exit"] = {"Exit this subsystem", Comm.cmd_exit},
}


--
-- Missile subsystem
--
Smoc = Host:new("smoc", "[SMOC]")
Smoc.motd = [[

_______________VERY IMPORTANT READ_______________

Ministry of Weapons replacing all peanut brittle warheads with bubble
gum, as mandated by FLD-1492 "Fearless Grandson Peanut Allergy".  Launch
capacity will be reduced until conversions are complete.

::: FLD-711 Restricted Distribution :::   wopr:xigeh-lydut-vinax
]]
Smoc.authcode = "CPE-1704-TKS"
Smoc.inventory = {
   "ready", "offline", "offline", "ready",
   "offline", "offline", "offline", "offline",
   "offline", "offline", "FileNotFound", "ready",
   [-1] = "program_invocation_short_name^@realm^@",
   [-2] = "^@^@^@^@^@^@^@^@^@^R^@^@^@3",
   [-3] = "<FF><FF><FF><FF>%L<C9>^D^Hhx^@^@^@",
   [-4] = "^D^H^G^P^@^@P",
   [-5] = "<EC>^P<8B>=<EC><C9>^D^H<C7>",
   [-6] = "WVS<83><EC>\<8B>E^L<8B><8B>U^P",
   [-7] = "Y^@^@" .. Smoc.authcode .. "^@get_launch_trigger^@",
   [-8] = "^@",
   [-9] = "^@",
   [-10] = "^@",
   [-11] = "^@wopr:xipar-canit-zimyx^@",
   [-12] = "^@",
   [-13] = "^@",
   [-14] = "^@",
   [-15] = "^@",
   [-16] = "^@",
}

function Smoc:login()
   if self:get("nuked") then
      Wopr:login("*** LINK DOWN\n*** CONNECTION REFUSED")
   else
      Host.login(self)
   end
end

function Smoc:cmd_exit(argv)
   Wopr:login()
end

function Smoc:cmd_status(argv)
   local n = tonumber(argv[2])
   if not n then
      local k, v, max
      local ready = 0
      for k,v in ipairs(self.inventory) do
         if (v == "ready") then
            ready = ready + 1
         end
         max = k
      end
      self:writeln(("%d total, %d ready"):format(max, ready))
      self:writeln("Use \"status #\" to check status of individual missiles")
   else
      self:writeln(("---- Missile #%d Summary ----"):format(n))
      self:writeln("Type: SS-256 SCUMM")
      self:writeln("Location: Fearless Missile Silo #1 (-44.76,-120.66)")
      self:writeln("Status: " .. (self.inventory[n] or "(null)"))
   end
end   

function Smoc:cmd_authorize(argv)
   if not Wopr:get("launch") then
      self:writeln("ERROR: Launch trigger disabled.")
   elseif (argv[2] ~= self.authcode) then
      self:writeln("Invalid authorization code.")
   else
      self:writeln("Authorization code accepted.")
      self:writeln("wopr:xocec-lifoz-gasyx")
      self:set("auth")
   end
end

function Smoc:cmd_launch(argv)
   local n = tonumber(argv[2])
   local lat = tonumber(argv[3])
   local lon = tonumber(argv[4])

   if Wopr:attcon() > 1 then
      self:writeln("ERROR: Missiles may only be launched during times of war.")
   elseif not self:get("auth") then
      self:writeln("ERROR: Not authorized")
   elseif (not n) then
      self:writeln("Usage: launch # LAT LONG")
   elseif (not lat) or (not lon) then
      self:writeln("ERROR: Invalid coordinates supplied")
   elseif (self.inventory[n] == "offline") then
      self:writeln("ERROR: Missile currently off-line")
   elseif (n < 1) then
      self:writeln("ERROR: No such missile")
   else
      self:writeln(("Launching to (%f,%f)..."):format(lat, lon))
      self:writeln("wopr:xubif-hikig-mocox")
      if (lat ~= -44.76) or (lon ~= -120.66) then
         self:writeln("ERROR: No propulsion system attached")
      elseif (self.inventory[n] ~= "FileNotFound") then
         self:writeln("ERROR: Triggering device not installed")
      else
         self:set("nuked")
         Wopr:login("Detonating warhead...\nwopr:xoroz-hymaz-fivex wopr:xufov-sugig-zecox wopr:xocem-dabal-fisux wopr:xufez-dofas-tyvyx\s*** CONNECTION TERMINATED")
      end
   end
end

Smoc:add_commands{
   ["status"] = {"Check missile status", Smoc.cmd_status},
   ["launch"] = {"Launch missile", Smoc.cmd_launch},
   ["authorize"] = {"Set authorization code", Smoc.cmd_authorize},
   ["exit"] = {"Exit to WOPR", Smoc.cmd_exit},
}


function main()
   if (not f["s"]) or (f["s"] == "") then
      Login:login()
   else
      local h = hosts_by_name[get("host")] or Login
      txt, prompt = h:handle_request(f["v"] or "")
   end
end

function err(msg)
   reply("", "A>", msg .. "  wopr:xosov-tenoh-nebox\n\n" .. debug.traceback())
end

xpcall(main, err)
