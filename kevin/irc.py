#! /usr/bin/env python3

import asynchat
import asyncore
import socket
import sys
import traceback
import time

class IRCHandler(asynchat.async_chat):
    """IRC Server connection.

    This is the one you want to derive your connection classes from.

    """

    debug = False
    heartbeat_interval = 1              # seconds per heartbeat

    def __init__(self, host=None, nick=None, gecos=None):
        asynchat.async_chat.__init__(self)
        self.line = b''
        self.timers = []
        self.last_heartbeat = 0
        self.set_terminator(b'\r\n')
        if host:
            self.open_connection(host, nick, gecos)

    def dbg(self, msg):
        if self.debug:
            print(msg)

    def open_connection(self, host, nick, gecos):
        self.nick = nick
        self.gecos = gecos
        self.host = host
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect(host)

    def handle_connect(self):
        self.write(['NICK', self.nick])
        self.write(['USER', self.nick, '+iw', self.nick], self.gecos)

    def connect(self, host):
        self.waiting = False
        asynchat.async_chat.connect(self, host)

    def heartbeat(self):
        """Invoke all timers."""

        if not self.timers:
            return
        timers, self.timers = self.timers, []
        now = time.time()
        for t, cb in timers:
            if t > now:
                self.timers.append((t, cb))
            else:
                cb()

    def add_timer(self, secs, callback):
        """After secs seconds, call callback"""
        self.timers.append((time.time() + secs, callback))

    def readable(self):
        """Called by asynchat to see if we're readable.

        We hook our heartbeat in here.
        """

        now = time.time()
        if now > self.last_heartbeat + self.heartbeat_interval:
            self.heartbeat()
            self.last_heartbeat = now

        if self.connected:
            return asynchat.async_chat.readable(self)
        else:
            return False

    def collect_incoming_data(self, data):
        """Called by asynchat when data arrives"""
        self.line += data

    def found_terminator(self):
        """Called by asynchat when it finds the terminating character.
        """
        line = self.line.decode('utf-8')
        self.line = b''
        self.parse_line(line)

    def write(self, args, text=None):
        """Send out an IRC command

        This function helps to prevent you from shooting yourself in the
        foot, by forcing you to send commands that are in a valid format
        (although it doesn't check the validity of the actual commands).

        As we all know, IRC commands take the form

          :COMMAND ARG1 ARG2 ARG3 ... :text string

        where 'text string' is optional.  Well, that's exactly how this
        function works.  Args is a list of length at least one, and text
        string is a string.

          write(['PRIVMSG', nick], 'Hello 12')

        will send

          PRIVMSG nick :Hello 12

        """

        cmdstr = ' '.join(args)
        if text:
            cmdstr = '%s :%s' % (cmdstr, text)
        self.dbg('-> %s' % cmdstr)
        try:
            line = '%s\n' % cmdstr
            self.send(line.encode('utf-8'))
        except socket.error:
            pass


    def parse_line(self, line):
        """Parse a server-provided line

        This does all the magic of parsing those ill-formatted IRC
        messages.  It will also decide if a PRIVMSG or NOTICE is using
        CTCP (the client-to-client protocol, which by convention is any
        of the above messages with ^A on both ends of the text.

        This function goes on to invoke self.eval_triggers on the parsed
        data like this:

          self.eval_triggers(operation, arguments, text)

        where operation and text are strings, and arguments is a list.

        It returns the same tuple (op, args, text).

        """

        if (line[0] == ':'):
            with_uname = 1
            line = line [1:]
        else:
            with_uname = 0
        try:
            [args, text] = line.split(' :', 1)
            args = args.split()
        except ValueError:
            args = line.split()
            text = ''
        if (with_uname != 1):
            op = args[0]
        elif ((args[1] in ["PRIVMSG", "NOTICE"]) and
              (text and (text[0] == '\001') and (text[-1] == '\001'))):
            op = "C" + args[1]
            text = text[1:-1]
        else:
            op = args[1]
        self.dbg("<- %s %s %s" % (op, args, text))
        self.handle(op, args, text)
        return (op, args, text)


    def handle(self, op, args, text):
        """Take action on a server message

        Right now, just invokes

          self.do_[op](args, text)

        where [op] is the operation passed in.

        This is a good method to overload if you want a really advanced
        client supporting bindings.

        """
        try:
            method = getattr(self, "do_" + lower(op))
        except AttributeError:
            self.dbg("Unhandled: %s" % (op, args, text))
            return
        method(args, text)


class Recipient:
    """Abstract recipient object"""

    def __init__(self, interface, name):
        self._interface = interface
        self._name = name

    def __repr__(self):
        return 'Recipient(%s)' % self.name()

    def name(self):
        return self._name

    def is_channel(self):
        return False

    def write(self, cmd, addl):
        """Write a raw IRC command to our interface"""

        self._interface.write(cmd, addl)

    def cmd(self, cmd, text):
        """Send a command to ourself"""

        self.write([cmd, self._name], text)

    def msg(self, text):
        """Tell the recipient something"""

        self.cmd("PRIVMSG", text)

    def notice(self, text):
        """Send a notice to the recipient"""

        self.cmd("NOTICE", text)

    def ctcp(self, command, text):
        """Send a CTCP command to the recipient"""

        return self.msg("\001%s %s\001" % (command.upper(), text))

    def act(self, text):
        """Send an action to the recipient"""

        return self.ctcp("ACTION", text)

    def cnotice(self, command, text):
        """Send a CTCP notice to the recipient"""

        return self.notice("\001%s %s\001" % (command.upper(), text))

class Channel(Recipient):
    def __repr__(self):
        return 'Channel(%s)' % self.name()

    def is_channel(self):
        return True

class User(Recipient):
    def __init__(self, interface, name, user, host, op=False):
        Recipient.__init__(self, interface, name)
        self.user = user
        self.host = host
        self.op = op

    def __repr__(self):
        return 'User(%s, %s, %s)' % (self.name(), self.user, self.host)

def recipient(interface, name):
    if name[0] in ["&", "#"]:
        return Channel(interface, name)
    else:
        return User(interface, name, None, None)

class SmartIRCHandler(IRCHandler):
    """This is like the IRCHandler, except it creates Recipient objects
    for IRC messages.  The intent is to make it easier to write stuff
    without knowledge of the IRC protocol.

    """

    def recipient(self, name):
        return recipient(self, name)

    def err(self, exception):
        if self.debug:
            traceback.print_exception(*exception)

    def handle(self, op, args, text):
        """Parse more, creating objects and stuff

        makes a call to self.handle_op(sender, forum, addl)

        sender is always a Recipient object; if you want to reply
        privately, you can send your reply to sender.

        forum is a Recipient object corresponding with the forum over
        which the message was carried.  For user-to-user PRIVMSG and
        NOTICE commands, this is the same as sender.  For those same
        commands sent to a channel, it is the channel.  Thus, you can
        always send a reply to forum, and it will be sent back in an
        appropriate manner (ie. the way you expect).

        addl is a tuple, containing additional information which might
        be relelvant.  Here's what it will contain, based on the server
        operation:

          op       | addl
          ---------+----------------
          PRIVMSG  | text of the message
          NOTICE   | text of the notice
          CPRIVMSG | CTCP command,  text of the command
          CNOTICE  | CTCP response, text of the response
          KICK *   | victim of kick, kick text
          MODE *   | all mode args
          JOIN *   | empty
          PART *   | empty
          QUIT     | quit message
          PING     | ping text
          NICK !   | old nickname
          others   | all arguments; text is last element

        * The forum in these items is the channel to which the action
          pertains.
        ! The sender for the NICK command is the *new* nickname.  This
          is so you can send messages to the sender object and they'll
          go to the right place.
        """

        try:
            sender = User(self, *unpack_nuhost(args))
        except ValueError:
            sender = None
        forum = None
        addl = ()

        if op in ("PRIVMSG", "NOTICE"):
            # PRIVMSG ['neale!~user@127.0.0.1', 'PRIVMSG', '#hydra'] firebot, foo
            # PRIVMSG ['neale!~user@127.0.0.1', 'PRIVMSG', 'firebot'] firebot, foo
            try:
                if args[2][0] in '#&':
                    forum = self.recipient(args[2])
                else:
                    forum = sender
                addl = (text,)
            except IndexError:
                addl = (text, args[1])
        elif op in ("CPRIVMSG", "CNOTICE"):
            forum = self.recipient(args[2])
            splits = text.split(" ")
            if splits[0] == "DCC":
                op = "DC" + op
                addl = (splits[1],) + tuple(splits[2:])
            else:
                addl = (splits[0],) + tuple(splits[1:])
        elif op in ("KICK",):
            forum = self.recipient(args[2])
            addl = (self.recipient(args[3]), text)
        elif op in ("MODE",):
            forum = self.recipient(args[2])
            addl = args[3:]
        elif op in ("JOIN", "PART"):
            try:
                forum = self.recipient(args[2])
            except IndexError:
                forum = self.recipient(text)
        elif op in ("QUIT",):
            addl = (text,)
        elif op in ("PING", "PONG"):
            # PING ['PING'] us.boogernet.org.
            # PONG ['irc.foonet.com', 'PONG', 'irc.foonet.com'] 1114199424
            addl = (text,)
        elif op in ("NICK",):
            # NICK ['brad!~brad@10.168.2.33', 'NICK'] bradaway
            #
            # The sender is the new nickname here, in case you want to
            # send something to the sender.

            # Apparently there are two different standards for this
            # command.
            if text:
                sender = self.recipient(text)
            else:
                sender = self.recipient(args[2])
            addl = (unpack_nuhost(args)[0],)
        elif op in ("INVITE",):
            # INVITE [u'pflarr!~pflarr@www.clanspum.net', u'INVITE', u'gallium', u'#mysterious']
            # INVITE [u'pflarr!~pflarr@www.clanspum.net', u'INVITE', u'gallium'] #mysterious
            if len(args) > 3:
                forum = self.recipient(args[3])
            else:
                forum = self.recipient(text)
        else:
            try:
                int(op)
            except ValueError:
                self.dbg("WARNING: unknown server code: %s" % op)
            addl = tuple(args[3:]) + (text,)

        try:
            self.handle_cooked(op, sender, forum, addl)
        except SystemExit:
            raise
        except:
            self.err(sys.exc_info())

    def handle_cooked(self, op, sender, forum, addl):
        try:
            func = getattr(self, 'cmd_' + op.lower())
        except AttributeError:
            self.unhandled(op, sender, forum, addl)
            return
        func(sender, forum, addl)

    def cmd_ping(self, sender, forum, addl):
        self.write(['PONG'], addl[0])

    def unhandled(self, op, sender, forum, addl):
        """Handle all the stuff that had no handler.

        This is a special handler in that it also gets the server code
        as the first argument.

        """

        self.dbg("unhandled: %s" % ((op, sender, forum, addl),))


class Bot(SmartIRCHandler):
    """A simple bot.

    This automatically joins the channels you pass to the constructor,
    tries to use one of the nicks provided, and reconnects if it gets
    booted.  You can use this as a base for more sophisticated bots.

    """

    def __init__(self, host, nicks, gecos, channels):
        self.nicks = nicks
        self.channels = channels
        self.waiting = True
        self._spool = []
        SmartIRCHandler.__init__(self, host, nicks[0], gecos)

    def despool(self, target, lines):
        """Slowly despool a bunch of lines to a target

        Since the IRC server will block all output if we send it too
        fast, use this to send large multi-line responses.

        """

        self._spool.append((target, list(lines)))

    def heartbeat(self):
        SmartIRCHandler.heartbeat(self)

        # Despool data
        if self._spool:
            # Take the first one on the queue, and put it on the end
            which = self._spool[0]
            del self._spool[0]
            self._spool.append(which)

            # Despool a line
            target, lines = which
            if lines:
                line = lines[0]
                target.msg(line)
                del lines[0]
            else:
                self._spool.remove(which)

    def announce(self, text):
        for c in self.channels:
            self.write(['PRIVMSG', c], text)

    def err(self, exception):
        SmartIRCHandler.err(self, exception)
        self.announce('*bzert*')

    def cmd_001(self, sender, forum, addl):
        for c in self.channels:
            self.write(['JOIN'], c)

    def writable(self):
        if not self.waiting:
            return asynchat.async_chat.writable(self)
        else:
            return False

    def write(self, *args):
        SmartIRCHandler.write(self, *args)

    def close(self, final=False):
        SmartIRCHandler.close(self)
        if not final:
            self.dbg("Connection closed, reconnecting...")
            self.waiting = True
            self.connected = 0
            # Wait a bit and reconnect
            self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
            self.add_timer(23, lambda : self.connect(self.host))

    def handle_close(self):
        self.close()


##
## Miscellaneous IRC functions
##

def unpack_nuhost(nuhost):
    """Unpack nick!user@host

    Frequently, the first argument in a server message is in
    nick!user@host format.  You can just pass your whole argument list
    to this function and get back a tuple containing:

      (nick, user, host)

    """

    try:
        [nick, uhost] = nuhost[0].split('!', 1)
        [user, host] = uhost.split('@', 1)
    except ValueError:
        raise ValueError("not in nick!user@host format")
    return (nick, user, host)

def run_forever(timeout=2.0):
    """Run your clients forever.

    Just a handy front-end to asyncore.loop, so you don't have to import
    asyncore yourself.

    """

    asyncore.loop(timeout)
