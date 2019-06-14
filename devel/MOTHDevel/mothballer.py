#!/usr/bin/env python3

import argparse
import binascii
import hashlib
import io
import json
import logging
from . import moth
import os
import shutil
import tempfile
import zipfile
import random

SEEDFN = "SEED"


def write_kv_pairs(ziphandle, filename, kv):
    """ Write out a sorted map to file
    :param ziphandle: a zipfile object
    :param filename: The filename to write within the zipfile object
    :param kv:  the map to write out
    :return:
    """
    filehandle = io.StringIO()
    for key in sorted(kv.keys()):
        if isinstance(kv[key], list):
            for val in kv[key]:
                filehandle.write("%s %s\n" % (key, val))
        else:
            filehandle.write("%s %s\n" % (key, kv[key]))
    filehandle.seek(0)
    ziphandle.writestr(filename, filehandle.read())


def escape(s):
    return s.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;')


def build_category(categorydir, outdir):
    category_seed = random.getrandbits(32)

    categoryname = os.path.basename(categorydir.strip(os.sep))
    zipfilename = os.path.join(outdir, "%s.mb" % categoryname)
    logging.info("Building {} from {}".format(zipfilename, categorydir))

    if os.path.exists(zipfilename):
        # open and gather some state
        existing = zipfile.ZipFile(zipfilename, 'r')
        try:
            category_seed = int(existing.open(SEEDFN).read().strip())
        except Exception:
            pass
        existing.close()
    logging.debug("Using PRNG seed {}".format(category_seed))

    zipfileraw = tempfile.NamedTemporaryFile(delete=False)
    mothball = package(categoryname, categorydir, category_seed)
    shutil.copyfileobj(mothball, zipfileraw)
    zipfileraw.close()
    shutil.move(zipfileraw.name, zipfilename)


# Returns a file-like object containing the contents of the new zip file
def package(categoryname, categorydir, seed):
    zfraw = io.BytesIO()
    zf = zipfile.ZipFile(zfraw, 'x')
    zf.writestr("category_seed.txt", str(seed))

    cat = moth.Category(categorydir, seed)
    mapping = {}
    answers = {}
    summary = {}
    for puzzle in cat:
        logging.info("Processing point value {}".format(puzzle.points))

        hashmap = hashlib.sha1(str(seed).encode('utf-8'))
        hashmap.update(str(puzzle.points).encode('utf-8'))
        puzzlehash = hashmap.hexdigest()

        mapping[puzzle.points] = puzzlehash
        answers[puzzle.points] = puzzle.answers
        summary[puzzle.points] = puzzle.summary

        puzzledir = os.path.join("content", puzzlehash)
        for fn, f in puzzle.files.items():
            payload = f.stream.read()
            zf.writestr(os.path.join(puzzledir, fn), payload)

        obj = puzzle.package()
        zf.writestr(os.path.join(puzzledir, 'puzzle.json'), json.dumps(obj))

    write_kv_pairs(zf, 'map.txt', mapping)
    write_kv_pairs(zf, 'answers.txt', answers)
    write_kv_pairs(zf, 'summaries.txt', summary)

    # clean up
    zf.close()
    zfraw.seek(0)
    return zfraw

def main(args=None):
    if args is None:
        import sys  # pragma: nocover
        args = sys.argv  # pragma: nocover

    parser = argparse.ArgumentParser(description='Build a category package')
    parser.add_argument('outdir', help='Output directory')
    parser.add_argument('categorydirs', nargs='+', help='Directory of category source')
    args = parser.parse_args(args)

    logging.basicConfig(level=logging.DEBUG)

    outdir = os.path.abspath(args.outdir)
    for categorydir in args.categorydirs:
        categorydir = os.path.abspath(categorydir)
        build_category(categorydir, outdir)

if __name__ == '__main__':
    main()  # pragma: nocover
