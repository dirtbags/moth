#! /usr/bin/python

import zipfile
import pyexiv2
import os
import random
import shutil
import tempfile
import time

basedir = 'src'

jpegs = os.listdir(basedir)
jpegs.append(jpegs[0])
jpegs.append(jpegs[1])
jpl = len(jpegs)

payload = open('payload.zip').read()
pll = len(payload)

chunksize = pll / jpl
chunks = []
while payload:
    chunk = pyexiv2.StringToUndefined(payload[:chunksize])
    chunks.append(chunk)
    payload = payload[chunksize:]

date_time = (2009, 8, 20, 9, 15)
seconds = 12
ofiles = []
for fn in jpegs:
    src = open(os.path.join(basedir, fn))
    dst = tempfile.NamedTemporaryFile(prefix='img', suffix='.jpg')
    shutil.copyfileobj(src, dst)
    dst.flush()

    # Write exif chunk
    chunk = chunks.pop(0)
    i = pyexiv2.Image(dst.name)
    i.readMetadata()
    i['Exif.Image.0x1663'] = chunk
    i.writeMetadata()

    timestamp = date_time + (seconds,)
    zinfo = zipfile.ZipInfo(os.path.basename(dst.name), timestamp)
    zinfo.external_attr = 0644 << 16
    ofiles.append((dst.name, dst, zinfo))
    seconds += 1

ofiles.sort()

zip = zipfile.ZipFile('out.zip', 'w')
for _, f, zinfo in ofiles:
    f.seek(0)
    zip.writestr(zinfo, f.read())

print('whew!')
