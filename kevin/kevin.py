#! /usr/bin/env python3

import os
import optparse
import asynchat
import socket
import asyncore
from urllib.parse import quote_plus as quote

from ctf import irc
from ctf.flagd import Flagger

nobody = '\002[nobody]\002'

class Kevin(irc.Bot):
    def __init__(self, host, flagger, tokens, victims):
        irc.Bot.__init__(self, host,
                         ['kevin', 'kev', 'kevin_', 'kev_', 'kevinm', 'kevinm_'],
                         'Kevin',
                         ['+kevin'])
        self.flagger = flagger
        self.tokens = tokens
        self.victims = victims
        self.affiliation = {}

    def cmd_001(self, sender, forum, addl):
        self.write(['OPER', 'bot', 'BottyMcBotpants'])
        irc.Bot.cmd_001(self, sender, forum, addl)

    def cmd_JOIN(self, sender, forum, addl):
        if sender.name == self.nick:
            self.tell_flag(forum)

    def cmd_381(self, sender, forum, addl):
        # You are now an IRC Operator
        if self.nick != 'kevin':
            self.write(['KILL', 'kevin'], 'You are not kevin.  I am kevin.')
            self.write(['NICK', 'kevin'])

    def err(self, exception):
        """Save the traceback for later inspection"""
        irc.Bot.err(self, exception)
        t,v,tb = exception
        info = []
        while 1:
            info.append('%s:%d(%s)' % (os.path.basename(tb.tb_frame.f_code.co_filename),
                                       tb.tb_lineno,
                                       tb.tb_frame.f_code.co_name))
            tb = tb.tb_next
            if not tb:
                break
        del tb                          # just to be safe
        infostr = '[' + '] ['.join(info) + ']'
        self.last_tb = '%s %s %s' % (t, v, infostr)
        print(self.last_tb)

    def tell_flag(self, forum):
        forum.msg('%s has the flag.' % (self.flagger.flag or nobody))

    def cmd_PRIVMSG(self, sender, forum, addl):
        text = addl[0]
        if text.startswith('!'):
            parts = text[1:].split(' ', 1)
            cmd = parts[0].lower()
            if len(parts) > 1:
                args = parts[1]
            else:
                args = None
            if cmd.startswith('r'):
                # Register
                who = sender.name()
                if args:
                    self.affiliation[who] = args
                team = self.affiliation.get(who, nobody)
                forum.msg('%s is playing for %s' % (who, team))
            elif cmd.startswith('e'):
                # Embrace
                forum.ctcp('ACTION', 'hugs %s' % sender.name())
            elif cmd.startswith('f'):
                # Flag
                self.tell_flag(forum)
            elif cmd.startswith('h'):
                # Help
                forum.msg('Goal: Obtain a token with social engineering.')
                forum.msg('Commands: !help, !flag, !register [TEAM], !claim TOKEN, !victims, !embrace')
            elif cmd.startswith('c') and args:
                # Claim
                sn = sender.name()
                team = self.affiliation.get(sn)
                token = quote(args, safe='')
                fn = os.path.join(self.tokens, token)
                if not team:
                    forum.msg('%s: register first (!register TEAM).' % sn)
                elif self.flagger.flag == team:
                    forum.msg('%s: Greedy, greedy.' % sn)
                elif not os.path.exists(fn):
                    forum.msg('%s: Token does not exist (possibly already claimed).' % sn)
                else:
                    os.unlink(fn)
                    self.flagger.set_flag(team)
                    self.tell_flag(forum)
            elif cmd.startswith('v'):
                # Victims
                # Open the file each time, so it can change
                try:
                    for line in open(self.victims):
                        forum.msg(line.strip())
                except IOError:
                    forum.msg('There are no victims!')
            elif cmd == 'traceback':
                forum.msg(self.last_tb or 'No traceback')

def main():
    p = optparse.OptionParser()
    p.add_option('-t', '--tokens', dest='tokens', default='./tokens',
                 help='Directory containing tokens')
    p.add_option('-v', '--victims', dest='victims', default='victims.txt',
                 help='File containing victims information')
    p.add_option('-i', '--ircd', dest='ircd', default='localhost',
                 help='IRC server to connect to')
    p.add_option('-f', '--flagd', dest='flagd', default='localhost',
                 help='Flag server to connect to')
    p.add_option('-p', '--password', dest='password',
                 default='kevin:::7db3e44d53d4a466f8facd7b7e9aa2b7',
                 help='Flag server password')
    p.add_option('-c', '--channel', dest='channel', 
                 help='Channel to join')
    opts, args = p.parse_args()

    f = Flagger(opts.flagd, opts.password.encode('utf-8'))
    k = Kevin((opts.ircd, 6667), f, opts.tokens, opts.victims)
    irc.run_forever()

if __name__ == '__main__':
    main()
