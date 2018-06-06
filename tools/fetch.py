#! /usr/bin/python3

import requests
import zipfile

instance = "foundry"

url = "https://puzzles.cyberfire.training/{}/".format(instance)
url = url.rstrip("/")

r = requests.get(url + "/puzzles.json")
puzzles = r.json()

zf = zipfile.ZipFile("/tmp/{}.zip".format(instance), "w")
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
            furl="{}/{}/{}/{}".format(url, cat, dn, fn)
            data = requests.get(furl).content
            zf.writestr(path, data)
