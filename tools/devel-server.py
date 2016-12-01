#!/usr/bin/env python3

# To pick up any changes to this file without restarting anything:
#     while true; do ./tools/devel-server.py --once; done
# It's kludgy, but it gets the job done.
# Feel free to make it suck less, for example using the `tcpserver` program.

import glob
import html
import http.server
import io
import mistune
import moth
import os
import pathlib
import shutil
import socketserver
import sys
import traceback

try:
    from http.server import HTTPStatus
except ImportError:
    class HTTPStatus:
        OK = (200, 'OK', 'Request fulfilled, document follows')
        NOT_FOUND = (404, 'Not Found', 'Nothing matches the given URI')
        INTERNAL_SERVER_ERROR = (500, 'Internal Server Error', 'Server got itself in trouble')

sys.dont_write_bytecode = True

# XXX: This will eventually cause a problem. Do something more clever here.
seed = 1

def page(title, body):
    return """<!DOCTYPE html>
<html>
  <head>
    <title>{title}</title>
    <link rel="stylesheet" href="/files/src/www/res/style.css">
  </head>
  <body>
    <h1>{title}</h1>
    <div id="preview" class="terminal">
      {body}
    </div>
  </body>
</html>""".format(title=title, body=body)


def mdpage(body):
    try:
        title, _ = body.split('\n', 1)
    except ValueError:
        title = "Result"
    title = title.lstrip("#")
    title = title.strip()
    return page(title, mistune.markdown(body, escape=False))


class ThreadingServer(socketserver.ThreadingMixIn, http.server.HTTPServer):
    pass


class MothHandler(http.server.SimpleHTTPRequestHandler):
    puzzles_dir = "puzzles"

    def handle_one_request(self):
        try:
            super().handle_one_request()
        except:
            tbtype, value, tb = sys.exc_info()
            tblist = traceback.format_tb(tb, None) + traceback.format_exception_only(tbtype, value)
            payload = ("Traceback (most recent call last)\n" +
                    "".join(tblist[:-1]) +
                    tblist[-1]).encode('utf-8')
            self.send_response(HTTPStatus.INTERNAL_SERVER_ERROR)
            self.send_header("Content-Type", "text/plain; charset=utf-8")
            self.send_header("Content-Length", payload)
            self.end_headers()
            self.wfile.write(payload)
                    
    def do_GET(self):
        if self.path == "/":
            self.serve_front()
        elif self.path.startswith("/puzzles"):
            self.serve_puzzles()
        elif self.path.startswith("/files"):
            self.serve_file(self.translate_path(self.path))
        else:
            self.send_error(HTTPStatus.NOT_FOUND, "File not found")

    def translate_path(self, path):
        if path.startswith('/files'):
            path = path[7:]
        return super().translate_path(path)

    def serve_front(self):
        body = """
MOTH Development Server Front Page
====================

Yo, it's the front page.
There's stuff you can do here:

* [Available puzzles](/puzzles)
* [Raw filesystem view](/files/)
* [Documentation](/files/docs/)
* [Instructions](/files/docs/devel-server.md) for using this server

If you use this development server to run a contest,
you are a fool.
"""
        payload = mdpage(body).encode('utf-8')
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", len(payload))
        self.end_headers()
        self.wfile.write(payload)

    def serve_puzzles(self):
        body = io.StringIO()
        path = self.path.rstrip('/')
        parts = path.split("/")
        title = None
        fpath = None
        points = None
        cat = None
        puzzle = None

        try:
            fpath = os.path.join(self.puzzles_dir, parts[2])
            points = int(parts[3])
        except:
            pass

        if fpath:
            cat = moth.Category(fpath, seed)
        if points:
            puzzle = cat.puzzle(points)

        if not cat:
            title = "Puzzle Categories"
            body.write("<ul>")
            for i in sorted(glob.glob(os.path.join(self.puzzles_dir, "*", ""))):
                bn = os.path.basename(i.strip('/\\'))
                body.write('<li><a href="/puzzles/{}">puzzles/{}/</a></li>'.format(bn, bn))
            body.write("</ul>")
        elif not puzzle:
            # List all point values in a category
            title = "Puzzles in category `{}`".format(parts[2])
            body.write("<ul>")
            for points in cat.pointvals:
                body.write('<li><a href="/puzzles/{cat}/{points}/">puzzles/{cat}/{points}/</a></li>'.format(cat=parts[2], points=points))
            body.write("</ul>")
        elif len(parts) == 4:
            # Serve up a puzzle
            title = "{} puzzle {}".format(parts[2], parts[3])
            body.write("<h2>Body</h2>")
            body.write(puzzle.html_body())
            body.write("<h2>Files</h2>")
            body.write("<ul>")
            for name in puzzle.files:
                body.write('<li><a href="/puzzles/{cat}/{points}/{filename}">{filename}</a></li>'
                            .format(cat=parts[2], points=puzzle.points, filename=name))
            body.write("</ul>")
            body.write("<h2>Answers</h2>")
            body.write("<ul>")
            assert puzzle.answers, 'No answers defined'
            for a in puzzle.answers:
                body.write("<li><code>{}</code></li>".format(html.escape(a)))
            body.write("</ul>")
            body.write("<h2>Author</h2><p>{}</p>".format(puzzle.author))
            body.write("<h2>Summary</h2><p>{}</p>".format(puzzle.summary))
            if puzzle.logs:
                body.write("<h2>Debug Log</h2>")
                body.write('<ul class="log">')
                for l in puzzle.logs:
                    body.write("<li>{}</li>".format(html.escape(l)))
                body.write("</ul>")
        elif len(parts) == 5:
            # Serve up a puzzle file
            try:
                pfile = puzzle.files[parts[4]]
            except KeyError:
                self.send_error(HTTPStatus.NOT_FOUND, "File not found. Did you add it to the Files: header or puzzle.add_stream?")
                return
            ctype = self.guess_type(pfile.name)
            self.send_response(HTTPStatus.OK)
            self.send_header("Content-Type", ctype)
            self.end_headers()
            shutil.copyfileobj(pfile.stream, self.wfile)
            return

        payload = page(title, body.getvalue()).encode('utf-8')
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", len(payload))
        self.end_headers()
        self.wfile.write(payload)

    def serve_file(self, path):
        lastmod = None
        fspath = pathlib.Path(path)

        if fspath.is_dir():
            ctype = "text/html; charset=utf-8"
            payload = self.list_directory(path)
            # it sends headers but not body
            shutil.copyfileobj(payload, self.wfile)
        else:
            ctype = self.guess_type(path)
            try:
                payload = fspath.read_bytes()
            except OSError:
                self.send_error(HTTPStatus.NOT_FOUND, "File not found")
                return
            if path.endswith(".md"):
                ctype = "text/html; charset=utf-8"
                content = mdpage(payload.decode('utf-8'))
                payload = content.encode('utf-8')
            try:
                fs = fspath.stat()
                lastmod = self.date_time_string(fs.st_mtime)
            except:
                pass

        self.send_response(HTTPStatus.OK)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", len(payload))
        if lastmod:
            self.send_header("Last-Modified", lastmod)
        self.end_headers()
        self.wfile.write(payload)


def run(address=('localhost', 8080), once=False):
    httpd = ThreadingServer(address, MothHandler)
    print("=== Listening on http://{}:{}/".format(address[0], address[1]))
    if once:
        httpd.handle_request()
    else:
        httpd.serve_forever()

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description="MOTH puzzle development server")
    parser.add_argument('--puzzles', default='puzzles',
                        help="Directory containing your puzzles")
    parser.add_argument('--once', default=False, action='store_true',
                        help="Serve one page, then exit. For debugging the server.")
    args = parser.parse_args()
    MothHandler.puzzles_dir = args.puzzles
    run(once=args.once)
