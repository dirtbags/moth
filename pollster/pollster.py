#!/usr/bin/env python3

import os
import re
import sys
import time
import socket
import urllib.request

DEBUG         = True
POLL_INTERVAL = 2
IP_DIR        = 'iptest/'
REPORT_PATH   = 'iptest/pollster.html'

def socket_poll(ip, port, msg):
	''' Connect via socket to the specified (ip, port), send
	the specified msg and return the response or None if something
	went wrong. '''
	sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	
	try:
		sock.connect((ip, port))
	except Exception as e:
		return None
	
	sock.send(msg)
	resp = sock.recv(1024)
	if len(resp) == 0:
		return None
	
	sock.close()

	resp = resp.decode('utf-8')
	return resp

# PUT POLLS FUNCTIONS HERE
#  Each function should take an IP address and return a team name or None
#  if (a) the service is not up, (b) it doesn't return a valid team name.

def poll_fingerd(ip):
	''' Poll the fingerd service. '''
	resp = socket_poll(ip, 79, b'flag\n')
	if resp is None:
		return None
	return resp.strip('\r\n')

def poll_noted(ip):
	''' Poll the noted service. '''
	resp = socket_poll(ip, 4000, b'rflag\n')
	if resp is None:
		return None
	return resp.strip('\r\n')

def poll_catcgi(ip):
	''' Poll the cat.cgi web service. '''
	url = urllib.request.urlopen('http://%s/cat.cgi/flag' % ip)
	data = url.read()
	if len(data) == 0:
		return None
	return data.strip('\r\n')

def poll_tftpd(ip):
	''' Poll the ftp service. '''
	sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
	sock.connect((ip, 69))

	sock.send(b'\x00\x01' + b'flag' + b'\x00' + b'octet' + b'\x00')
	resp = sock.recv(1024)
	if len(resp) <= 5:
		return None
	
	sock.close()

	return resp[4:].decode('utf-8').strip('\r\n')
	
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

	# check that IP_DIR is there, exit if it isn't
	if not os.path.isdir(IP_DIR):
		sys.stderr.write('directory %s does not exist or is not readable\n' % IP_DIR)
		sys.exit(1)

	# gather the list of IPs to poll
	ips = os.listdir(IP_DIR)
	results = {}
	for ip in ips:

		# check file name format is ip
		if ip_re.match(ip) is None:
			continue

		#os.remove(os.path.join(IP_DIR, ip))

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

