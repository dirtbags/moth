#!/usr/bin/env python3

import argparse
import binascii
import glob
import hashlib
import io
import json
import logging
import moth
import os
import shutil
import string
import sys
import tempfile
import zipfile

def write_kv_pairs(ziphandle, filename, kv):
    """ Write out a sorted map to file
    :param ziphandle: a zipfile object
    :param filename: The filename to write within the zipfile object
    :param kv:  the map to write out
    :return:
    """
    filehandle = io.StringIO()
    for key in sorted(kv.keys()):
        if type(kv[key])  == type([]):
            for val in kv[key]:
                filehandle.write("%s: %s%s" % (key, val, os.linesep))
        else:
            filehandle.write("%s: %s%s" % (key, kv[key], os.linesep))
    filehandle.seek(0)
    ziphandle.writestr(filename, filehandle.read())

def build_category(categorydir, outdir):
    zipfileraw = tempfile.NamedTemporaryFile(delete=False)
    zf = zipfile.ZipFile(zipfileraw, 'x')

    category_seed = binascii.b2a_hex(os.urandom(20))
    puzzles_dict = {}
    secrets = {}

    categoryname = os.path.basename(categorydir.strip(os.sep))
    seedfn = os.path.join("category_seed.txt")
    zipfilename = os.path.join(outdir, "%s.zip" % categoryname)
    logging.info("Building {} from {}".format(zipfilename, categorydir))

    if os.path.exists(zipfilename):
        # open and gather some state
        existing = zipfile.ZipFile(zipfilename, 'r')
        try:
            category_seed = existing.open(seedfn).read().strip()
        except:
            pass
        existing.close()
    logging.debug("Using PRNG seed {}".format(category_seed))

    zf.writestr(seedfn, category_seed)

    cat = moth.Category(categorydir, category_seed)
    mapping = {}
    answers = {}
    summary = {}
    for puzzle in cat:
        logging.info("Processing point value {}".format(puzzle.points))

        hashmap = hashlib.sha1(category_seed)
        hashmap.update(str(puzzle.points).encode('utf-8'))
        puzzlehash = hashmap.hexdigest()
        
        mapping[puzzle.points] = puzzlehash
        answers[puzzle.points] = puzzle.answers
        summary[puzzle.points] = puzzle.summary

        puzzledir = os.path.join('content', puzzlehash)
        files = []
        for fn, f in puzzle.files.items():
            if f.visible:
                files.append(fn)
            payload = f.stream.read()
            zf.writestr(os.path.join(puzzledir, fn), payload)

        puzzledict = {
            'author': puzzle.author,
            'hashes': puzzle.hashes(),
            'files': files,
            'body': puzzle.html_body(),
        }
        puzzlejson = json.dumps(puzzledict)
        zf.writestr(os.path.join(puzzledir, 'puzzle.json'), puzzlejson)

    write_kv_pairs(zf, 'map.txt', mapping)
    write_kv_pairs(zf, 'answers.txt', answers)
    write_kv_pairs(zf, 'summaries.txt', summary)

    # clean up
    zf.close()

    shutil.move(zipfileraw.name, zipfilename)
    
   
if __name__ == '__main__':        
    parser = argparse.ArgumentParser(description='Build a category package')
    parser.add_argument('categorydirs', nargs='+', help='Directory of category source')
    parser.add_argument('outdir', help='Output directory')
    args = parser.parse_args()

    logging.basicConfig(level=logging.DEBUG)

    for categorydir in args.categorydirs:
        build_category(categorydir, args.outdir)

