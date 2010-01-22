#!/usr/bin/env python3

import os
import re
import io
import sys
import time
import socket
import traceback
import subprocess
import random
import http.client

from ctf import config
from ctf import pointscli

DEBUG	        = False
POLL_INTERVAL   = config.get('pollster', 'poll_interval')
IP_DIR	        = config.get('pollster', 'heartbeat_dir')
REPORT_PATH     = config.get('pollster', 'results')
SOCK_TIMEOUT    = config.get('pollster', 'poll_timeout')
POLL_IFACE      = config.get('pollster', 'poll_iface')
POLL_MAC_VENDOR = config.get('pollster', 'poll_mac_vendor')

class BoundHTTPConnection(http.client.HTTPConnection):
	''' http.client.HTTPConnection doesn't support binding to a particular
	address, which is something we need. '''
	
	def __init__(self, bindip, host, port=None, strict=None, timeout=socket._GLOBAL_DEFAULT_TIMEOUT):
		http.client.HTTPConnection.__init__(self, host, port, strict, timeout)
		self.bindip = bindip
	
	def connect(self):
		''' Connect to the host and port specified in __init__, but
		also bind first. '''
		self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		self.sock.bind((self.bindip, 0))
		self.sock.settimeout(self.timeout)
		self.sock.connect((self.host, self.port))

		if self._tunnel_host:
			self._tunnel()

def random_mac():
	''' Set a random mac on the poll interface. '''
	mac = ':'.join([POLL_MAC_VENDOR] + ['%02x' % random.randint(0,255) for i in range(3)])
	retcode = subprocess.call(('ifconfig', POLL_IFACE, 'hw', 'ether', mac)) 

def dhcp_request():
	''' Request a new IP on the poll interface. '''
	retcode = subprocess.call(('dhclient', POLL_IFACE))

def get_ip():
	''' Return the IP of the poll interface. '''
	path = os.path.join(os.getcwd(), 'get_ip.sh')
	p = subprocess.Popen((path, POLL_IFACE), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
	(out, err) = p.communicate()
	return out.strip(b'\r\n').decode('utf-8')

def socket_poll(srcip, ip, port, msg, prot, max_recv=1):
	''' Connect via socket to the specified <ip>:<port> using the
	specified <prot>, send the specified <msg> and return the
	response or None if something went wrong. <max_recvs> specifies
	how many times to read from the socket (defaults to once). '''

	# create a socket
	try:
		sock = socket.socket(socket.AF_INET, prot)
	except Exception as e:
		print('pollster: create socket failed (%s)' % e)
		traceback.print_exc()
		return None

	sock.bind((srcip, 0))
	sock.settimeout(SOCK_TIMEOUT)

	# connect
	try:
		sock.connect((ip, port))
	except socket.timeout as e:
		print('pollster: attempt to connect to %s:%d timed out (%s)' % (ip, port, e))
		traceback.print_exc()
		return None
	except Exception as e:
		print('pollster: attempt to connect to %s:%d failed (%s)' % (ip, port, e))
		traceback.print_exc()
		return None

	# send something
	sock.send(msg)

	# get a response
	resp = []
	try:
		# read from the socket until <max_recv> responses or read,
		# a timeout occurs, the socket closes, or some other exception
		# is raised
		for i in range(max_recv):
			data = sock.recv(1024)
			if len(data) == 0:
				break
			resp.append(data)

	except socket.timeout as e:
		print('pollster: timed out waiting for a response from %s:%d (%s)' % (ip, port, e))
		traceback.print_exc()
	except Exception as e:
		print('pollster: receive from %s:%d failed (%s)' % (ip, port, e))
		traceback.print_exc()

	sock.close()

	if len(resp) == 0:
		return None

	return b''.join(resp)

# PUT POLLS FUNCTIONS HERE
#  Each function should take an IP address and return a team name or None
#  if (a) the service is not up, (b) it doesn't return a valid team name.

def poll_fingerd(srcip, ip):
	''' Poll the fingerd service. Returns None or a team name. '''
	resp = socket_poll(srcip, ip, 79, b'flag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip(b'\r\n')

def poll_noted(srcip, ip):
	''' Poll the noted service. Returns None or a team name. '''
	resp = socket_poll(srcip, ip, 4000, b'rflag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip(b'\r\n')

def poll_catcgi(srcip, ip):
	''' Poll the cat.cgi web service. Returns None or a team name. '''
	
	try:
		conn = BoundHTTPConnection(srcip, ip, timeout=SOCK_TIMEOUT)
		conn.request('GET', '/cat.cgi/flag')
	except:
		return None

	resp = conn.getresponse()
	if resp.status != 200:
		conn.close()
		return None
	
	data = resp.read()
	conn.close()
	return data.strip(b'\r\n')

def poll_tftpd(srcip, ip):
	''' Poll the tftp service. Returns None or a team name. '''
	resp = socket_poll(srcip, ip, 69, b'\x00\x01' + b'flag' + b'\x00' + b'octet' + b'\x00', socket.SOCK_DGRAM)
	if resp is None:
		return None

	if len(resp) <= 5:
		return None

	resp = resp.split(b'\n')[0]

	# ack
	_ = socket_poll(srcip, ip, 69, b'\x00\x04' + resp[2:4], socket.SOCK_DGRAM, 0)

	return resp[4:].strip(b'\r\n')

# PUT POLL FUNCTIONS IN HERE OR THEY WONT BE POLLED
POLLS = {
	'fingerd'  : poll_fingerd,
	'noted'	   : poll_noted,
	'catcgi'   : poll_catcgi,
	'tftpd'	   : poll_tftpd,
}

ip_re = re.compile('(\d{1,3}\.){3}\d{1,3}')
poll_no = 0
# loop forever
while True:

	random_mac()
	dhcp_request()

	srcip = get_ip()

	t_start = time.time()

	# gather the list of IPs to poll
	ips = os.listdir(IP_DIR)

	out = io.StringIO()
	out.write(config.start_html('Team Service Availability'))
	for ip in ips:
		# check file name format is ip
		if ip_re.match(ip) is None:
			continue

		# remove the file
		try:
			os.remove(os.path.join(IP_DIR, ip))
		except Exception as e:
			print('pollster: could not remove %s' % os.path.join(IP_DIR, ip))
			traceback.print_exc()

		results = {}

		if DEBUG is True:
			print('ip: %s' % ip)

		if out is not None:
			out.write('<h2>%s</h2>\n' % ip)
			out.write('<table class="pollster">\n<thead><tr><td>Service Name</td></td>')
			out.write('<td>Flag Holder</td></tr></thead>\n')

		# perform polls
		for service,func in POLLS.items():
			try:
				team = func(srcip, ip).decode('utf-8')
				if len(team) == 0:
					team = 'dirtbags'
			except:
				team = 'dirtbags'

			if DEBUG is True:
				print('\t%s - %s' % (service, team))

			if out is not None:
				out.write('<tr><td>%s</td><td>%s</td>\n' % (service, team))

			pointscli.submit('svc.' + service, team, 1)

		if out is not None:
			out.write('</table>\n')

	if DEBUG is True:
		print('+-----------------------------------------+')

	out.write('<p>Poll number: %d</p>' % poll_no)
	poll_no += 1

	t_end = time.time()
	exec_time = int(t_end - t_start)
	sleep_time = POLL_INTERVAL - exec_time

	if out is not None:
		out.write(config.end_html())

	open(REPORT_PATH, 'w').write(out.getvalue())

	# sleep until its time to poll again
	time.sleep(sleep_time)

