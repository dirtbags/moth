#!/usr/bin/env python3
"""Module responsible for compiling Moth puzzles into mothballs"""

import argparse
import hashlib
import io
import json
import logging
import os
import shutil
import tempfile
import zipfile
import random

from . import moth

SEEDFN = "SEED"


def write_kv_pairs(ziphandle, filename, data):
    """ Write out a sorted map to file
    :param ziphandle: a zipfile object
    :param filename: The filename to write within the zipfile object
    :param kv:  the map to write out
    :return:
    """
    filehandle = io.StringIO()
    for key in sorted(data.keys()):
        if isinstance(data[key], list):
            for val in data[key]:
                filehandle.write("%s %s\n" % (key, val))
        else:
            filehandle.write("%s %s\n" % (key, data[key]))
    filehandle.seek(0)
    ziphandle.writestr(filename, filehandle.read())


def escape(instr):
    """Escape common HTML sequences
    :param instr: string to escape
    :return: the escaped string
    """
    return instr.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;')


def build_category(categorydir, outdir):
    """Build a category and save it as a mothball
    :param categorydir: the path to the directory containing category Moth files
    :param outdir: the path to the directory to which generated mothballs will be written
    :return:
    """
    category_seed = random.getrandbits(32)

    categoryname = os.path.basename(categorydir.strip(os.sep))
    zipfilename = os.path.join(outdir, "%s.mb" % categoryname)
    logging.info("Building %s from %s", zipfilename, categorydir)

    if os.path.exists(zipfilename):
        # open and gather some state
        existing = zipfile.ZipFile(zipfilename, 'r')
        try:
            category_seed = int(existing.open(SEEDFN).read().strip())
        except (ValueError, KeyError):  # If the seed is malformed or doesn't exist
            pass
        existing.close()
    logging.debug("Using PRNG seed %s", category_seed)

    zipfileraw = tempfile.NamedTemporaryFile(delete=False)
    mothball = package(categorydir, category_seed)
    shutil.copyfileobj(mothball, zipfileraw)
    zipfileraw.close()
    shutil.move(zipfileraw.name, zipfilename)


def package(categorydir, seed):
    """Return a file-like object containing the contents of the new zip file
    :param categorydir: The path to the directory containing category Moth files
    :param seed: A seed value used to initialize the PSRNG used by some puzzles
    :return: A file-like object containing the contents of the new zip file
    """
    # pylint: disable=too-many-locals
    zfraw = io.BytesIO()
    with zipfile.ZipFile(zfraw, 'x') as mothball:
        mothball.writestr("category_seed.txt", str(seed))

        cat = moth.Category(categorydir, seed)
        mapping = {}
        answers = {}
        summary = {}
        for puzzle in cat:
            logging.info("Processing point value %s", puzzle.points)

            hashmap = hashlib.sha1(str(seed).encode('utf-8'))
            hashmap.update(str(puzzle.points).encode('utf-8'))
            puzzlehash = hashmap.hexdigest()

            mapping[puzzle.points] = puzzlehash
            answers[puzzle.points] = puzzle.answers
            summary[puzzle.points] = puzzle.summary

            puzzledir = os.path.join("content", puzzlehash)
            for filename, file_obj in puzzle.files.items():
                payload = file_obj.stream.read()
                mothball.writestr(os.path.join(puzzledir, filename), payload)

            obj = puzzle.package()
            mothball.writestr(os.path.join(puzzledir, 'puzzle.json'), json.dumps(obj))

        write_kv_pairs(mothball, 'map.txt', mapping)
        write_kv_pairs(mothball, 'answers.txt', answers)
        write_kv_pairs(mothball, 'summaries.txt', summary)

    # clean up
    zfraw.seek(0)
    return zfraw


def main(args=None):
    """Main interface"""
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
