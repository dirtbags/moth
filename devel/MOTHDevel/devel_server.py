#!/usr/bin/python3
"""The MOTH development server"""

import cgi
import cgitb
from http import HTTPStatus
import http.server
import json
import mimetypes
import logging
import os
import pathlib
import random
import shutil
import sys
import posixpath

from . import moth
from . import mothballer
from . import parse

sys.dont_write_bytecode = True  # Don't write .pyc files

try:
    ThreadingHTTPServer = http.server.ThreadingHTTPServer  # pylint: disable=invalid-name
except AttributeError:
    import socketserver

    class ThreadingHTTPServer(socketserver.ThreadingMixIn, http.server.HTTPServer):
        """Mocked-up ThreadingHTTPServer

        This is used if the running version of Python doesn't have a
        native ThreatdingHTTPServer
        """
        daemon_threads = True


class MothServer(ThreadingHTTPServer):
    """The MOTH Development server"""
    def __init__(self, server_address, RequestHandlerClass):
        super().__init__(server_address, RequestHandlerClass)
        self.args = {}


class MothRequestHandler(http.server.SimpleHTTPRequestHandler):
    """A basic request handler for the MOTH development server"""
    endpoints = []

    def __init__(self, request, client_address, server):
        self.directory = str(server.args["theme_dir"])
        try:
            super().__init__(request, client_address, server, directory=server.args["theme_dir"])
        except TypeError:
            super().__init__(request, client_address, server)

        self.req = None
        self.seed = None
        self.fields = None
        self.path = None

    # Backport from Python 3.7
    def translate_path(self, path):
        # I guess we just hope that some other thread doesn't call getcwd
        getcwd = os.getcwd
        os.getcwd = lambda: self.directory
        ret = super().translate_path(path)
        os.getcwd = getcwd
        return ret

    def get_puzzle(self):
        """Retrieve puzzle information, based on GET parameters"""
        category = self.req.get("cat")
        points = int(self.req.get("points"))
        catpath = str(self.server.args["puzzles_dir"].joinpath(category))
        cat = moth.Category(catpath, self.seed)
        puzzle = cat.puzzle(points)
        return puzzle

    def handle_answer(self):

        """Handle requests for the answer endpoint"""

        for field_name in ("cat", "points", "answer"):
            self.req[field_name] = self.fields.getfirst(field_name)

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
        """Handle requests for the puzzlelist endpoint"""
        puzzles = {
            "__devel__": [[0, ""]],
        }

        for puzzle in self.server.args["puzzles_dir"].glob("*"):
            if not puzzle.is_dir() or puzzle.match(".*"):
                continue

            category_name = puzzle.parts[-1]
            cat = moth.Category(str(puzzle), self.seed)
            puzzles[category_name] = [[i, str(i)] for i in cat.pointvals()]
            puzzles[category_name].append([0, ""])

        if len(puzzles) <= 1:
            logging.warning("No directories found matching %s/*", self.server.args["puzzles_dir"])

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(puzzles).encode("utf-8"))
    endpoints.append(('/{seed}/puzzles.json', handle_puzzlelist))

    def handle_puzzle(self):
        """Handle requests to the puzzle endpoint"""
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
        """Handle requests for files in puzzles"""
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
        content_type, content_encoding = mimetypes.guess_type(file.name)
        if content_type is not None:
            if content_encoding is not None:
                content_type += ";encoding=" + content_encoding

            self.send_header("Content-Type", content_type)

        self.end_headers()
        shutil.copyfileobj(file.stream, self.wfile)

    endpoints.append(("/{seed}/content/{cat}/{points}/{filename}", handle_puzzlefile))

    def handle_mothballer(self):
        """Handle requests for the mothballer endpoint"""
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
        else:
            self.send_response(200)
            self.send_header("Content-Type", "application/octet_stream")
            self.end_headers()
            shutil.copyfileobj(mothball, self.wfile)

    endpoints.append(("/{seed}/mothballer/{cat}.mb", handle_mothballer))

    def handle_index(self):
        """Handle requests for the dev server index page"""
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
        """Handle requests for MOTH theme files"""
        self.path = "/" + self.req.get("path", "")
        super().do_GET()

    endpoints.append(("/{seed}/", handle_theme_file))
    endpoints.append(("/{seed}/{path}", handle_theme_file))

    def do_GET(self):
        """Handle HTTP GET requests"""
        self.fields = cgi.FieldStorage(
            fp=self.rfile,
            headers=self.headers,
            environ={
                "REQUEST_METHOD": self.command,
                "CONTENT_TYPE": self.headers["Content-Type"],
            },
        )

        for pattern, function in self.endpoints:
            result = parse.parse(pattern, self.path)
            if result:
                self.req = result.named
                seed = self.req.get("seed", "random")
                if seed == "random":
                    self.seed = random.getrandbits(32)
                else:
                    self.seed = int(seed)
                return function(self)

        return super().do_GET()

    def do_POST(self):  # pylint: disable=invalid-name
        """Handle HTTP POST requests"""
        self.do_GET()

    def do_HEAD(self):
        """Handle HTTP HEAD requests"""
        self.send_error(
            HTTPStatus.NOT_IMPLEMENTED,
            "Unsupported method (%r)" % self.command,
        )


def main():
    """Main function handler"""
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

    server = MothServer((addr, port), MothRequestHandler)
    server.args["base_url"] = args.base
    server.args["puzzles_dir"] = pathlib.Path(args.puzzles)
    server.args["theme_dir"] = args.theme

    logging.info("Listening on %s:%d", addr, port)
    server.serve_forever()


if __name__ == "__main__":
    main()
