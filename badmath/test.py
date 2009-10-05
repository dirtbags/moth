import Pi, irc

pi = Pi.pi(('irc.lanl.gov', 6667), '')
irc.run_forever()
