#!/usr/bin/python3

import asyncio
import glob
import html
from aiohttp import web
import io
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

sys.dont_write_bytecode = True  # Don't write .pyc files

def mkseed():
    return bytes(random.choice(b'abcdef0123456789') for i in range(40))

class Page:
    def __init__(self, title, depth=0):
        self.title = title
        if depth:
            self.base = "/".join([".."] * depth)
        else:
            self.base = "."
        self.body = io.StringIO()
        self.scripts = []
        
    def add_script(self, path):
        self.scripts.append(path)
        
    def write(self, s):
        self.body.write(s)
        
    def text(self):
        ret = io.StringIO()
        ret.write("<!DOCTYPE html>\n")
        ret.write("<html>\n")
        ret.write("  <head>\n")
        ret.write("    <title>{}</title>\n".format(self.title))
        ret.write("    <link rel=\"stylesheet\" href=\"{}/files/www/res/style.css\">\n".format(self.base))
        for s in self.scripts:
            ret.write("    {}\n".format(s))
        ret.write("  </head>\n")
        ret.write("  <body>\n")
        ret.write("    <h1>{}</h1>\n".format(self.title))
        ret.write("    <div id=\"preview\" class=\"terminal\">\n")
        ret.write(self.body.getvalue())
        ret.write("    </div>\n")
        ret.write("  </body>\n")
        ret.write("</html>\n")
        return ret.getvalue()
        
    def response(self, request):
        return web.Response(text=self.text(), content_type="text/html")

async def handle_front(request):
    p = Page("Devel Server", 0)
    p.write("<p>Yo, it's the front page!</p>")
    p.write("<ul>")
    p.write("<li><a href=\"puzzles/\">Available puzzles</a></li>")
    p.write("<li><a href=\"files/\">Raw filesystem view</a></li>")
    p.write("<li><a href=\"https://github.com/dirtbags/moth/tree/master/docs\">Documentation</a></li>")
    p.write("<li><a href=\"https://github.com/dirtbags/moth/blob/master/docs/devel-server.md\"Instructions</a> for using this server")
    p.write("</ul>")
    p.write("<p>If you use this development server to run a contest, you are a fool.</p>")
    return p.response(request)

async def handle_puzzlelist(request):
    p = Page("Puzzle Categories", 1)
    p.write("<ul>")
    for i in sorted(glob.glob(os.path.join(request.app["puzzles_dir"], "*", ""))):
        bn = os.path.basename(i.strip('/\\'))
        p.write('<li><a href="{}/">puzzles/{}/</a></li>'.format(bn, bn))
    p.write("</ul>")
    return p.response(request)

async def handle_category(request):
    seed = request.query.get("seed", mkseed())
    category = request.match_info.get("category")
    cat = moth.Category(os.path.join(request.app["puzzles_dir"], category), seed)
    p = Page("Puzzles in category {}".format(category), 2)
    p.write("<ul>")
    for points in cat.pointvals():
        p.write('<li><a href="{points}/">puzzles/{category}/{points}/</a></li>'.format(category=category, points=points))
    p.write("</ul>")
    return p.response(request)

async def handle_puzzle(request):
    seed = request.query.get("seed", mkseed())
    category = request.match_info.get("category")
    points = int(request.match_info.get("points"))
    cat = moth.Category(os.path.join(request.app["puzzles_dir"], category), seed)
    puzzle = cat.puzzle(points)

    p = Page("{} puzzle {}".format(category, points), 3)
    for s in puzzle.scripts:
        p.add_script(s)
    p.write("<h2>Body</h2>")
    p.write("<div id='body' style='border: solid 1px silver;'>")
    p.write(puzzle.html_body())
    p.write("</div>")
    p.write("<h2>Files</h2>")
    p.write("<ul>")
    for name,puzzlefile in sorted(puzzle.files.items()):
        if puzzlefile.visible:
            visibility = ''
        else:
            visibility = '(unlisted)'
        p.write('<li><a href="{filename}">{filename}</a> {visibility}</li>'
                    .format(cat=category,
                            points=puzzle.points,
                            filename=name,
                            visibility=visibility))
    p.write("</ul>")
    p.write("<h2>Answers</h2>")
    p.write("<p>Input box (for scripts): <input id='answer' name='a'>")
    p.write("<ul>")
    assert puzzle.answers, 'No answers defined'
    for a in puzzle.answers:
        p.write("<li><code>{}</code></li>".format(html.escape(a)))
    p.write("</ul>")
    p.write("<h2>Authors</h2><p>{}</p>".format(', '.join(puzzle.get_authors())))
    p.write("<h2>Summary</h2><p>{}</p>".format(puzzle.summary))
    if puzzle.logs:
        p.write("<h2>Debug Log</h2>")
        p.write('<ul class="log">')
        for l in puzzle.logs:
            p.write("<li>{}</li>".format(html.escape(l)))
        p.write("</ul>")
        
    return p.response(request)

async def handle_puzzlefile(request):
    seed = request.query.get("seed", mkseed())
    category = request.match_info.get("category")
    points = int(request.match_info.get("points"))
    filename = request.match_info.get("filename")
    cat = moth.Category(os.path.join(request.app["puzzles_dir"], category), seed)
    puzzle = cat.puzzle(points)

    try:
        file = puzzle.files[filename]
    except KeyError:
        return web.Response(status=404)
    
    resp = web.Response()
    resp.content_type, _ = mimetypes.guess_type(file.name)
    # This is the line where I decided Go was better than Python at multiprocessing
    # You should be able to chain the puzzle file's output to the async output,
    # without having to block. But if there's a way to do that, it certainly
    # isn't documented anywhere.
    resp.body = file.stream.read()
    return resp


if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description="MOTH puzzle development server")
    parser.add_argument(
        '--puzzles', default='puzzles',
        help="Directory containing your puzzles"
    )
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
    app["puzzles_dir"] = args.puzzles
    app["base_url"] = args.base
    app.router.add_route("GET", "/", handle_front)
    app.router.add_route("GET", "/puzzles/", handle_puzzlelist)
    app.router.add_route("GET", "/puzzles/{category}/", handle_category)
    app.router.add_route("GET", "/puzzles/{category}/{points}/", handle_puzzle)
    app.router.add_route("GET", "/puzzles/{category}/{points}/{filename}", handle_puzzlefile)
    app.router.add_static("/files/", mydir, show_index=True)
    web.run_app(app, host=addr, port=port)
