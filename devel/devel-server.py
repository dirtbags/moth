#!/usr/bin/python3

import cgitb
import html
import cgi
import http.server
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
import parse
import urllib.parse
import posixpath
from http import HTTPStatus


sys.dont_write_bytecode = True  # Don't write .pyc files


class MothServer(socketserver.ForkingMixIn, http.server.HTTPServer):
    def __init__(self, server_address, RequestHandlerClass):
        super().__init__(server_address, RequestHandlerClass)
        self.args = {}


class MothRequestHandler(http.server.SimpleHTTPRequestHandler):
    endpoints = []
    
    def __init__(self, request, client_address, server):
        self.directory = str(server.args["theme_dir"])
        try:
            super().__init__(request, client_address, server, directory=server.args["theme_dir"])
        except TypeError:
            super().__init__(request, client_address, server)
        # Why can't they just use mimetypes?!

    # Why isn't this the default?!
    def guess_type(self, path):
        mtype, encoding = mimetypes.guess_type(path)
        if encoding:
            return "%s; encoding=%s" % (mtype, encoding)
        else:
            return mtype

    # Backport from Python 3.7
    def translate_path(self, path):
        # I guess we just hope that some other thread doesn't call getcwd
        getcwd = os.getcwd
        os.getcwd = lambda: self.directory
        ret = super().translate_path(path)
        os.getcwd = getcwd
        return ret


    def get_puzzle(self):
        category = self.req.get("cat")
        points = int(self.req.get("points"))
        catpath = str(self.server.args["puzzles_dir"].joinpath(category))
        cat = moth.Category(catpath, self.seed)
        puzzle = cat.puzzle(points)
        return puzzle


    def handle_answer(self):
        for f in ("cat", "points", "answer"):
            self.req[f] = self.fields.getfirst(f)
        puzzle = self.get_puzzle()
        ret = {
            "status": "success",
            "data": {
               "short": "",
               "description": "Provided answer was not in list of answers"
            },
        }

        if self.req.get("answer") in puzzle.answers:
            ret["data"]["description"] = "Answer is correct"
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(ret).encode("utf-8"))
    endpoints.append(('/{seed}/answer', handle_answer))

    
    def handle_puzzlelist(self):
        puzzles = {
            "__devel__": [[0, ""]],
        }
        for p in self.server.args["puzzles_dir"].glob("*"):
            if not p.is_dir() or p.match(".*"):
                continue
            catName = p.parts[-1]
            cat = moth.Category(str(p), self.seed)
            puzzles[catName] = [[i, str(i)] for i in cat.pointvals()]
            puzzles[catName].append([0, ""])
        if len(puzzles) <= 1:
            logging.warning("No directories found matching {}/*".format(self.server.args["puzzles_dir"]))
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(puzzles).encode("utf-8"))
    endpoints.append(('/{seed}/puzzles.json', handle_puzzlelist))
    
    
    def handle_puzzle(self):
        puzzle = self.get_puzzle()

        obj = puzzle.package()
        obj["answers"] = puzzle.answers
        obj["hint"] = puzzle.hint
        obj["summary"] = puzzle.summary
        obj["logs"] = puzzle.logs

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(obj).encode("utf-8"))
    endpoints.append(('/{seed}/content/{cat}/{points}/puzzle.json', handle_puzzle))


    def handle_puzzlefile(self):
        puzzle = self.get_puzzle()

        try:
            file = puzzle.files[self.req["filename"]]
        except KeyError:
            self.send_error(
                HTTPStatus.NOT_FOUND,
                "File Not Found",
            )
            return

        self.send_response(200)
        self.send_header("Content-Type", mimetypes.guess_type(file.name))
        self.end_headers()
        shutil.copyfileobj(file.stream, self.wfile)
    endpoints.append(("/{seed}/content/{cat}/{points}/{filename}", handle_puzzlefile))
    

    def handle_mothballer(self):
        category = self.req.get("cat")
        
        try:
            catdir = self.server.args["puzzles_dir"].joinpath(category)
            mb = mothballer.package(category, catdir, self.seed)
        except Exception as ex:
            logging.exception(ex)
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=\"utf-8\"")
            self.end_headers()
            self.wfile.write(cgitb.html(sys.exc_info()))
            return
        
        self.send_response(200)
        self.send_header("Content-Type", "application/octet_stream")
        self.end_headers()
        shutil.copyfileobj(mb, self.wfile)
    endpoints.append(("/{seed}/mothballer/{cat}.mb", handle_mothballer))


    def handle_index(self):
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

        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.end_headers()
        self.wfile.write(body.encode('utf-8'))
    endpoints.append((r"/", handle_index))


    def handle_theme_file(self):
        self.path = "/" + self.req.get("path", "")
        super().do_GET()
    endpoints.append(("/{seed}/", handle_theme_file))
    endpoints.append(("/{seed}/{path}", handle_theme_file))


    def do_GET(self):
        self.fields = cgi.FieldStorage(
            fp=self.rfile,
            headers=self.headers,
            environ={
                "REQUEST_METHOD": self.command,
                "CONTENT_TYPE": self.headers["Content-Type"],
            },
        )

        url = urllib.parse.urlparse(self.path)
        for pattern, function in self.endpoints:
            result = parse.parse(pattern, url.path)
            if result:
                self.req = result.named
                seed = self.req.get("seed", "random")
                if seed == "random":
                    self.seed = random.getrandbits(32)
                else:
                    self.seed = int(seed)
                return function(self)
        super().do_GET()

    def do_POST(self):
        self.do_GET()

    def do_HEAD(self):
        self.send_error(
            HTTPStatus.NOT_IMPLEMENTED,
            "Unsupported method (%r)" % self.command,
        )


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
    parser.add_argument(
        "-v", "--verbose",
        action="count",
        default=1,  # Leave at 1, for now, to maintain current default behavior
        help="Include more verbose logging. Use multiple flags to increase level",
    )
    args = parser.parse_args()
    parts = args.bind.split(":")
    addr = parts[0] or "0.0.0.0"
    port = int(parts[1])
    if args.verbose >= 2:
        log_level = logging.DEBUG
    elif args.verbose == 1:
        log_level = logging.INFO
    else:
        log_level = logging.WARNING
    
    logging.basicConfig(level=log_level)
    
    mimetypes.add_type("application/javascript", ".mjs")

    server = MothServer((addr, port), MothRequestHandler)
    server.args["base_url"] = args.base
    server.args["puzzles_dir"] = pathlib.Path(args.puzzles)
    server.args["theme_dir"] = args.theme
    

    logging.info("Listening on %s:%d", addr, port)
    server.serve_forever()
