#!/usr/bin/env python3

import os
import re
import sys
import time
import socket
import urllib.request

# TODO:
# special stops for http and tftp?
# get start time and end time of poll, sleep(60-exectime)
# what to do about exceptions
# no nested dicts
# scoring interface
# config interface
# html that uses the proper css

DEBUG         = True
POLL_INTERVAL = 2
IP_DIR        = 'iptest/'
REPORT_PATH   = 'iptest/pollster.html'
SOCK_TIMEOUT  = 0.5

def socket_poll(ip, port, msg, prot, max_recv=1):
	''' Connect via socket to the specified <ip>:<port> using the
	specified <prot>, send the specified <msg> and return the 
	response or None if something went wrong. <max_recvs> specifies
	how many times to read from the socket. '''

	# create a socket
	try:
		sock = socket.socket(socket.AF_INET, prot)
	except Exception as e:
		print('pollster: create socket failed')
		return None
	
	sock.settimeout(SOCK_TIMEOUT)

	# connect
	try:
		sock.connect((ip, port))
	except socket.timeout as e:
		print('pollster: attempt to connect to %s:%d timed out' % (ip, port))
		return None
	except Exception as e:
		print('pollster: attempt to connect to %s:%d failed' % (ip, port))
		return None

	# send something
	sock.send(msg)

	# get a response
	resp = ''
	try:
		# first read
		data = sock.recv(1024)
		resp += data.decode('utf-8')
		max_recv -= 1

		# remaining reads as necessary until timeout or socket closes
		while(len(data) > 0 and max_recv > 0):
			data = sock.recv(1024)
			resp += data.decode('utf-8')
			max_recv -= 1
		sock.close()
	except socket.timeout as e:
		print('pollster: timed out waiting for a response from %s:%d' % (ip, port))
	except Exception as e:
		print('pollster: receive from %s:%d failed' % (ip, port))
	
	if len(resp) == 0:
		return None

	return resp

# PUT POLLS FUNCTIONS HERE
#  Each function should take an IP address and return a team name or None
#  if (a) the service is not up, (b) it doesn't return a valid team name.

def poll_fingerd(ip):
	''' Poll the fingerd service. '''
	resp = socket_poll(ip, 79, b'flag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip('\r\n')

def poll_noted(ip):
	''' Poll the noted service. '''
	resp = socket_poll(ip, 4000, b'rflag\n', socket.SOCK_STREAM)
	if resp is None:
		return None
	return resp.strip('\r\n')

def poll_catcgi(ip):
	''' Poll the cat.cgi web service. '''
	request = bytes('GET /cat.cgi/flag HTTP/1.1\r\nHost: %s\r\n\r\n' % ip, 'ascii')
	resp = socket_poll(ip, 80, request, socket.SOCK_STREAM, 3)
	if resp is None:
		return None

	content = resp.split('\r\n\r\n')
	if len(content) < 3:
		return None

	content = content[1].split('\r\n')

	try:
		content_len = int(content[0])
	except Exception as e:
		return None
	
	if content_len <= 0:
		return None
	return content[1].strip('\r\n')

def poll_tftpd(ip):
	''' Poll the tftp service. '''
	resp = socket_poll(ip, 69, b'\x00\x01' + b'flag' + b'\x00' + b'octet' + b'\x00', socket.SOCK_DGRAM)
	if resp is None:
		return None
	
	if len(resp) <= 5:
		return None
	
	resp = resp.split('\n')[0]
	return resp[4:].strip('\r\n')
	
# PUT POLL FUNCTIONS IN HERE OR THEY WONT BE POLLED
POLLS = {
	'fingerd'  : poll_fingerd,
	'noted'    : poll_noted,
	'catcgi'   : poll_catcgi,
	'tftpd'    : poll_tftpd,
}

ip_re = re.compile('(\d{1,3}\.){3}\d{1,3}')

# loop forever
while(True):

	# gather the list of IPs to poll
	try:
		ips = os.listdir(IP_DIR)
	except Exception as e:
		print('pollster: could not list dir %s' % IP_DIR)

	results = {}
	for ip in ips:

		# check file name format is ip
		if ip_re.match(ip) is None:
			continue

		# remove the file
		#try:
		#	os.remove(os.path.join(IP_DIR, ip))
		#except Exception as e:
		#	print('pollster: could not remove %s' % os.path.join(IP_DIR, ip))

		results[ip] = {}

		if DEBUG is True:
			print('ip: %s' % ip)

		# perform polls
		for service,func in POLLS.items():
			team = func(ip)
			if team is None:
				team = 'dirtbags'

			results[ip][service] = team	

		if DEBUG is True:
			for k,v in results[ip].items():
				print('\t%s - %s' % (k,v))
		
	if DEBUG is True:
		print('+-----------------------------------------+')
	
	# allocate points

	# generate html report
	out = open(REPORT_PATH, 'w')
	out.write('<html>\n<title><head>Polling Results</head></title>\n')
	out.write('<body>\n<h1>Polling Results</h1>\n')

	for ip in results.keys():
		out.write('<h2>%s</h2>\n' % ip)
		out.write('<table>\n<thead><tr><td>Service Name</td></td>')
		out.write('<td>Flag Holder</td></tr></thead>\n')
		for service,flag_holder in results[ip].items():
			out.write('<tr><td>%s</td><td>%s</td>\n' % (service, flag_holder))
		out.write('</table>\n')
	
	out.write('</body>\n</html>\n')
	out.close()
	
	# sleep until its time to poll again
	time.sleep(POLL_INTERVAL)

