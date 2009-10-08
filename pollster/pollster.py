#!/usr/bin/env python3

import os
import re
import sys
import time
import socket
import traceback

from ctf import config
from ctf import pointscli

DEBUG         = False
POLL_INTERVAL = config.get('pollster', 'poll_interval')
IP_DIR        = config.get('pollster', 'heartbeat_dir')
REPORT_PATH   = config.get('pollster', 'results')
SOCK_TIMEOUT  = config.get('pollster', 'poll_timeout')


def socket_poll(ip, port, msg, prot, max_recv=1):
	''' Connect via socket to the specified <ip>:<port> using the
	specified <prot>, send the specified <msg> and return the
	response or None if something went wrong. <max_recvs> specifies
	how many times to read from the socket (default to once). '''

	# create a socket
	try:
		sock = socket.socket(socket.AF_INET, prot)
	except Exception as e:
		print('pollster: create socket failed (%s)' % e)
		traceback.print_exc()
		return None

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

def poll_fingerd(ip):
	''' Poll the fingerd service. Returns None or a team name. '''
	resp = socket_poll(ip, 79, b'flag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip(b'\r\n')

def poll_noted(ip):
	''' Poll the noted service. Returns None or a team name. '''
	resp = socket_poll(ip, 4000, b'rflag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip(b'\r\n')

def poll_catcgi(ip):
	''' Poll the cat.cgi web service. Returns None or a team name. '''
	request = bytes('GET /cat.cgi/flag HTTP/1.1\r\nHost: %s\r\n\r\n' % ip, 'ascii')
	resp = socket_poll(ip, 80, request, socket.SOCK_STREAM, 3)
	if resp is None:
		return None

	content = resp.split(b'\r\n\r\n')
	if len(content) < 3:
		return None

	content = content[1].split(b'\r\n')

	try:
		content_len = int(content[0])
	except Exception as e:
		return None

	if content_len <= 0:
		return None
	return content[1].strip(b'\r\n')

def poll_tftpd(ip):
	''' Poll the tftp service. Returns None or a team name. '''
	resp = socket_poll(ip, 69, b'\x00\x01' + b'flag' + b'\x00' + b'octet' + b'\x00', socket.SOCK_DGRAM)
	if resp is None:
		return None

	if len(resp) <= 5:
		return None

	resp = resp.split(b'\n')[0]

	# ack
	_ = socket_poll(ip, 69, b'\x00\x04' + resp[2:4], socket.SOCK_DGRAM, 0)

	return resp[4:].strip(b'\r\n')

# PUT POLL FUNCTIONS IN HERE OR THEY WONT BE POLLED
POLLS = {
	'fingerd'  : poll_fingerd,
	'noted'	   : poll_noted,
	'catcgi'   : poll_catcgi,
	'tftpd'	   : poll_tftpd,
}

ip_re = re.compile('(\d{1,3}\.){3}\d{1,3}')

# loop forever
while True:

	t_start = time.time()

	# gather the list of IPs to poll
	try:
		ips = os.listdir(IP_DIR)
	except Exception as e:
		print('pollster: could not list dir %s (%s)' % (IP_DIR, e))
		traceback.print_exc()

	try:
		os.remove(REPORT_PATH)
	except Exception as e:
		pass

	try:
		out = open(REPORT_PATH, 'w')
	except Exception as e:
		out = None

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
			team = func(ip).decode('utf-8')
			if team is None:
				team = 'dirtbags'

			if DEBUG is True:
				print('\t%s - %s' % (service, team))

			if out is not None:
				out.write('<tr><td>%s</td><td>%s</td>\n' % (service, team))

			pointscli.submit(service, team, 1)

		if out is not None:
			out.write('</table>\n')

	if DEBUG is True:
		print('+-----------------------------------------+')

	t_end = time.time()
	exec_time = int(t_end - t_start)
	sleep_time = POLL_INTERVAL - exec_time

	if out is not None:
		out.write(config.end_html())
		out.close()

	# sleep until its time to poll again
	time.sleep(sleep_time)

