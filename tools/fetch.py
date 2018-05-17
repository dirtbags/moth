#! /usr/bin/python3

import requests
import zipfile

url = "https://puzzles.cyberfire.training/foundry/"
url = url.rstrip("/")

r = requests.get(url + "/puzzles.json")
puzzles = r.json()

zf = zipfile.ZipFile("/tmp/foundry.zip", "w")
for cat, entries in puzzles.items():
    if cat == "wopr":
        continue
    
    for points, dn in entries:
        if points == 0:
            continue
        u = "{}/{}/{}/puzzle.json".format(url, cat, dn)

        print(u, points, dn)
        obj = requests.get(u).json()
        files = obj.get("files") + ["index.html"]
        
        for fn in files:
            path = "{}/{}/{}".format(cat, points, fn)
            data = requests.get(u).content
            zf.writestr(path, data)
