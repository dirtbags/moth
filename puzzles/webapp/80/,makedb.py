#!/usr/bin/env python2.6

import os
import sys
import sqlite3
import base64

# new db
if os.path.exists(',zomg.sqlite3'):
	os.remove(',zomg.sqlite3')
db = sqlite3.connect(',zomg.sqlite3')
cur = db.cursor()

# pics table
cur.execute('create table pics(id integer primary key, data blob)')
paths = os.listdir(',pics/')
for path in paths:
	f = open(os.path.join(',pics/', path), 'rb')
	data = f.read()
	f.close()
	encoded = base64.encodestring(data)
	html = '<img src="data:image/jpg;base64,%s"/>' % encoded
	cur.execute('insert into pics(data) values(?)', (html,))

# jokes table
cur.execute('create table jokes(id integer primary key, data text)')
paths = os.listdir(',jokes/')
for path in paths:
	f = open(os.path.join(',jokes/', path), 'r')
	html = f.read()
	f.close()
	cur.execute('insert into jokes(data) values(?)', (html,))

# key
cur.execute('create table key(id integer primary key, data text)')
for k in [None, None, None, None, None, 'dmW5f9P54e']:
	cur.execute('insert into key(data) values(?)', (k,))

# clean up
db.commit()
cur.close()
db.close()

