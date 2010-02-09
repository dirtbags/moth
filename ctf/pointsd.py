#! /usr/bin/env python3

import asyncore
import socket
import struct
import time
from . import points
from . import config

house = config.get('global', 'house_team')

class PointsServer(asyncore.dispatcher):
	''' Receive connections from client and passes them off to handler. '''

	def __init__(self, port=6667):
		asyncore.dispatcher.__init__(self)
		self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
		self.bind(('', port))
		self.listen(5)
		self.acked = set()
		self.outq = []

	def handle_accept(self):
		''' Accept a connection from a client and pass it to the handler. '''
		sock, addr = self.accept()
		clientip = addr[0]
		ClientHandler(sock, clientip)

class ClientHandler(asyncore.dispatcher):
	''' Handles talking to clients. '''

	def __init__(self, sock, clientip):
		asyncore.dispatcher.__init__(self, sock=sock)
		self.clientip = clientip
		self.store = points.Storage(fix=True)
		self.acked = set()
		self.outq = []

	def writable(self):
		''' If there is data in the queue, the socket is made writable. '''
		return bool(self.outq)
	
	def handle_write(self):
		''' Pop data from the queue and send it to the client. '''
		resp = self.outq.pop(0)
		self.send(resp)

		# conversation over
		self.close()
	
	def handle_read(self):
		''' Receive data from the client. '''
		now = int(time.time())
		data = self.recv(4096)

		# decode their message
		try:
			id, when, cat, team, score = points.decode_request(data)
		except ValueError as e:
			return self.respond(now, str(e))
		team = team or house

		# do points and send ACK
		if not ((self.clientip, id) in self.acked):
			if not (now - 2 < when <= now):
				return self.respond(id, 'Your clock is off')
			self.store.add((when, cat, team, score))
			self.acked.add((self.clientip, id))

		self.respond(id, 'OK')
	
	def respond(self, id, txt):
		''' Queue responses to the client. '''
		resp = points.encode_response(id, txt)
		self.outq.append(resp)

if __name__ == '__main__':
	server = PointsServer()
	asyncore.loop()

