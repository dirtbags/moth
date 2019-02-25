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

        
def get_puzzle(request, data=None):
    seed = get_seed(request)
    if not data:
        data = request.match_info
    category = data.get("cat")
    points = int(data.get("points"))
    filename = data.get("filename")
    cat = moth.Category(request.app["puzzles_dir"].joinpath(category), seed)
    puzzle = cat.puzzle(points)
    return puzzle

async def handle_answer(request):
    data = await request.post()
    puzzle = get_puzzle(request, data)
    ret = {
        "status": "success",
        "data": {
           "short": "",
           "description": "Provided answer was not in list of answers"
        },
    }
    
    if data.get("answer") in puzzle.answers:
        ret["data"]["description"] = "Answer is correct"
    return web.Response(
        content_type="application/json",
        body=json.dumps(ret),
    )
    

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
    category = request.match_info.get("cat")
    points = int(request.match_info.get("points"))
    cat = moth.Category(request.app["puzzles_dir"].joinpath(category), seed)
    puzzle = cat.puzzle(points)
    
    obj = puzzle.package()
    obj["answers"] = puzzle.answers
    obj["hint"] = puzzle.hint
    obj["summary"] = puzzle.summary
    obj["logs"] = puzzle.logs
    
    return web.Response(
        content_type="application/json",
        body=json.dumps(obj),
    )
    

async def handle_puzzlefile(request):
    seed = get_seed(request)
    category = request.match_info.get("cat")
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
    category = request.match_info.get("cat")
    
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
  <head>
    <title>Dev Server</title>
    <script>
// Skip trying to log in
sessionStorage.setItem("id", "devel-server")
    </script>
  </head>
  <body>
    <h1>Dev Server</h1>

    <p>
      Pick a seed:
    </p>
    <ul>
      <li><a href="{seed}/">{seed}</a>: a special seed I made just for you!</li>
      <li><a href="random/">random</a>: will use a different seed every time you load a page (could be useful for debugging)</li>
      <li>You can also hack your own seed into the URL, if you want to.</li>
    </ul>

    <p>
      Puzzles can be generated from Python code: these puzzles can use a random number generator if they want.
      The seed is used to create these random numbers.
    </p>
    
    <p>
      We like to make a new seed for every contest,
      and re-use that seed whenever we regenerate a category during an event
      (say to fix a bug).
      By using the same seed,
      we make sure that all the dynamically-generated puzzles have the same values
      in any new packages we build.
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
        fn = "index.html"
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
    app.router.add_route("*", "/{seed}/answer", handle_answer)
    app.router.add_route("*", "/{seed}/puzzles.json", handle_puzzlelist)
    app.router.add_route("GET", "/{seed}/content/{cat}/{points}/puzzle.json", handle_puzzle)
    app.router.add_route("GET", "/{seed}/content/{cat}/{points}/{filename}", handle_puzzlefile)
    app.router.add_route("GET", "/{seed}/mothballer/{cat}", handle_mothballer)
    app.router.add_route("GET", "/{seed}/{filename:.*}", handle_static)
    web.run_app(app, host=addr, port=port)
