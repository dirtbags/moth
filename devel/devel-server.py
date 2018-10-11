#!/usr/bin/python3

import asyncio
import cgitb
import html
from aiohttp import web
import io
import json
import mimetypes
import moth
import logging
import os
import pathlib
import random
import shutil
import socketserver
import sys
import traceback
import mothballer

sys.dont_write_bytecode = True  # Don't write .pyc files


def get_seed(request):
    seedstr = request.match_info.get("seed")
    if seedstr == "random":
        return random.getrandbits(32)
    else:
        return int(seedstr)


async def handle_puzzlelist(request):
    seed = get_seed(request)
    puzzles = {
        "__devel__": [[0, ""]],
    }
    for p in request.app["puzzles_dir"].glob("*"):
        if not p.is_dir() or p.match(".*"):
            continue
        catName = p.parts[-1]
        cat = moth.Category(p, seed)
        puzzles[catName] = [[i, str(i)] for i in cat.pointvals()]
        puzzles[catName].append([0, ""])
    if len(puzzles) <= 1:
        logging.warning("No directories found matching {}/*".format(request.app["puzzles_dir"]))
    return web.Response(
        content_type="application/json",
        body=json.dumps(puzzles),
    )


async def handle_puzzle(request):
    seed = get_seed(request)
    category = request.match_info.get("category")
    points = int(request.match_info.get("points"))
    cat = moth.Category(request.app["puzzles_dir"].joinpath(category), seed)
    puzzle = cat.puzzle(points)
    
    obj = puzzle.package()
    obj["answers"] = puzzle.answers
    obj["hint"] = puzzle.hint
    obj["summary"] = puzzle.summary
    
    return web.Response(
        content_type="application/json",
        body=json.dumps(obj),
    )
    

async def handle_puzzlefile(request):
    seed = get_seed(request)
    category = request.match_info.get("category")
    points = int(request.match_info.get("points"))
    filename = request.match_info.get("filename")
    cat = moth.Category(request.app["puzzles_dir"].joinpath(category), seed)
    puzzle = cat.puzzle(points)

    try:
        file = puzzle.files[filename]
    except KeyError:
        return web.Response(status=404)

    content_type, _ = mimetypes.guess_type(file.name)
    return web.Response(
        body=file.stream.read(), # Is there no way to pipe this, must we slurp the whole thing into memory?
        content_type=content_type,
    )


async def handle_mothballer(request):
    seed = get_seed(request)
    category = request.match_info.get("category")
    
    try:
        catdir = request.app["puzzles_dir"].joinpath(category)
        mb = mothballer.package(category, catdir, seed)
    except:
        body = cgitb.html(sys.exc_info())
        resp = web.Response(text=body, content_type="text/html")
        return resp
        
    mb_buf = mb.read()
    resp = web.Response(
        body=mb_buf,
        headers={"Content-Disposition": "attachment; filename={}.mb".format(category)},
        content_type="application/octet_stream",
    )
    return resp


async def handle_index(request):
    seed = random.getrandbits(32)
    body = """<!DOCTYPE html>
<html>
  <head><title>Dev Server</title></head>
  <body>
    <h1>Dev Server</h1>
    <p>
      You need to provide the contest seed in the URL.
      If you don't have a contest seed in mind,
      why not try <a href="{seed}/">{seed}</a>?
    </p>
    <p>
      If you are chaotic,
      you could even take your chances with a
      <a href="random/">random seed</a> for every HTTP request.
      This means generated files will get a different seed than the puzzle itself!
    </p>
  </body>
</html>
""".format(seed=seed)
    return web.Response(
        content_type="text/html",
        body=body,
    )


async def handle_static(request):
    themes = request.app["theme_dir"]
    fn = request.match_info.get("filename")
    if not fn:
        for fn in ("puzzle-list.html", "index.html"):
            path = themes.joinpath(fn)
            if path.exists():
                break
    else:
        path = themes.joinpath(fn)
    return web.FileResponse(path)


if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description="MOTH puzzle development server")
    parser.add_argument(
        '--puzzles', default='puzzles',
        help="Directory containing your puzzles"
    )
    parser.add_argument(
        '--theme', default='theme',
        help="Directory containing theme files")
    parser.add_argument(
        '--bind', default="127.0.0.1:8080",
        help="Bind to ip:port"
    )
    parser.add_argument(
        '--base', default="",
        help="Base URL to this server, for reverse proxy setup"
    )
    args = parser.parse_args()
    parts = args.bind.split(":")
    addr = parts[0] or "0.0.0.0"
    port = int(parts[1])
    
    logging.basicConfig(level=logging.INFO)
    
    mydir = os.path.dirname(os.path.dirname(os.path.realpath(sys.argv[0])))
    
    app = web.Application()
    app["base_url"] = args.base
    app["puzzles_dir"] = pathlib.Path(args.puzzles)
    app["theme_dir"] = pathlib.Path(args.theme)
    app.router.add_route("GET", "/", handle_index)
    app.router.add_route("GET", "/{seed}/puzzles.json", handle_puzzlelist)
    app.router.add_route("GET", "/{seed}/content/{category}/{points}/puzzle.json", handle_puzzle)
    app.router.add_route("GET", "/{seed}/content/{category}/{points}/{filename}", handle_puzzlefile)
    app.router.add_route("GET", "/{seed}/mothballer/{category}", handle_mothballer)
    app.router.add_route("GET", "/{seed}/{filename:.*}", handle_static)
    web.run_app(app, host=addr, port=port)
