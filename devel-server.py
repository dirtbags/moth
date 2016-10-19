#!/usr/bin/env python3

import cgi
import glob
import http.server
import mistune
import os
import pathlib
import puzzles
import socketserver
import sys
import traceback

try:
    from http.server import HTTPStatus
except ImportError:
    class HTTPStatus:
        NOT_FOUND = 404
        OK = 200

# XXX: This will eventually cause a problem. Do something more clever here.
seed = 1


def page(title, body):
    return """<!DOCTYPE html>
<html>
  <head>
    <title>{}</title>
    <link rel="stylesheet" href="/files/www/res/style.css">
  </head>
  <body>
    <div id="preview" class="terminal">
      {}
    </div>
  </body>
</html>""".format(title, body)


def mdpage(body):
    try:
        title, _ = body.split('\n', 1)
    except ValueError:
        title = "Result"
    title = title.lstrip("#")
    title = title.strip()
    return page(title, mistune.markdown(body))


class ThreadingServer(socketserver.ThreadingMixIn, http.server.HTTPServer):
    pass


class MothHandler(http.server.SimpleHTTPRequestHandler):
    def handle_one_request(self):
        try:
            super().handle_one_request()
        except:
            tbtype, value, tb = sys.exc_info()
            tblist = traceback.format_tb(tb, None) + traceback.format_exception_only(tbtype, value)
            page = ("# Traceback (most recent call last)\n" +
                    "    " +
                    "    ".join(tblist[:-1]) +
                    tblist[-1])
            self.serve_md(page)
            
                    
    def do_GET(self):
        if self.path == "/":
            self.serve_front()
        elif self.path.startswith("/puzzles"):
            self.serve_puzzles()
        elif self.path.startswith("/files"):
            self.serve_file()
        else:
            self.send_error(HTTPStatus.NOT_FOUND, "File not found")

    def translate_path(self, path):
        if path.startswith('/files'):
            path = path[7:]
        return super().translate_path(path)

    def serve_front(self):
        page = """
MOTH Development Server Front Page
====================

Yo, it's the front page.
There's stuff you can do here:

* [Available puzzles](/puzzles)
* [Raw filesystem view](/files/)
* [Documentation](/files/doc/)
* [Instructions](/files/doc/devel-server.md) for using this server

If you use this development server to run a contest,
you are a fool.
"""
        self.serve_md(page)

    def serve_puzzles(self):
        body = []
        path = self.path.rstrip('/')
        parts = path.split("/")
        #raise ValueError(parts)
        if len(parts) < 3:
            # List all categories
            body.append("# Puzzle Categories")
            for i in glob.glob(os.path.join("puzzles", "*", "")):
                body.append("* [{}](/{})".format(i, i))
            self.serve_md('\n'.join(body))
            return

        fpath = os.path.join("puzzles", parts[2])
        cat = puzzles.Category(fpath, seed)
        if len(parts) == 3:
            # List all point values in a category
            body.append("# Puzzles in category `{}`".format(parts[2]))
            for points in cat.pointvals:
                body.append("* [puzzles/{cat}/{points}](/puzzles/{cat}/{points}/)".format(cat=parts[2], points=points))
            self.serve_md('\n'.join(body))
            return

        pzl = cat.puzzle(int(parts[3]))
        if len(parts) == 4:
            body.append("# {} puzzle {}".format(parts[2], parts[3]))
            body.append("* Author: `{}`".format(pzl['author']))
            body.append("* Summary: `{}`".format(pzl['summary']))
            body.append('')
            body.append("## Body")
            body.append(pzl.body)
            body.append("## Answers")
            for a in pzl['answer']:
                body.append("* `{}`".format(a))
            body.append("")
            body.append("## Files")
            for pzl_file in pzl['files']:
                body.append("* [puzzles/{cat}/{points}/{filename}]({filename})"
                            .format(cat=parts[2], points=pzl.points, filename=pzl_file))
            self.serve_md('\n'.join(body))
            return
        elif len(parts) == 5:
            try:
                self.serve_puzzle_file(pzl['files'][parts[4]])
            except KeyError:
                self.send_error(HTTPStatus.NOT_FOUND, "File not found")
                return
        else:
            body.append("# Not Implemented Yet")
            self.serve_md('\n'.join(body))

    CHUNK_SIZE = 4096
    def serve_puzzle_file(self, file):
        """Serve a PuzzleFile object."""
        self.send_response(HTTPStatus.OK)
        self.send_header("Content-type", "application/octet-stream")
        self.send_header('Content-Disposition', 'attachment; filename="{}"'.format(file.name))
        if file.path is not None:
            fs = os.stat(file.path)
            self.send_header("Last-Modified", self.date_time_string(fs.st_mtime))

        # We're using application/octet stream, so we can send the raw bytes.
        self.end_headers()
        chunk = file.handle.read(self.CHUNK_SIZE)
        while chunk:
            self.wfile.write(chunk)
            chunk = file.handle.read(self.CHUNK_SIZE)

    def serve_file(self):
        if self.path.endswith(".md"):
            self.serve_md()
        else:
            super().do_GET()

    def serve_md(self, text=None):
        fspathstr = self.translate_path(self.path)
        fspath = pathlib.Path(fspathstr)
        if not text:
            try:
                text = fspath.read_text()
            except OSError:
                self.send_error(HTTPStatus.NOT_FOUND, "File not found")
                return None
        content = mdpage(text)

        self.send_response(HTTPStatus.OK)

        self.send_header("Content-type", "text/html; charset=utf-8")
        self.send_header("Content-Length", len(content))
        try:
            fs = fspath.stat()
            self.send_header("Last-Modified", self.date_time_string(fs.st_mtime))
        except:
            pass
        self.end_headers()
        self.wfile.write(content.encode('utf-8'))


def run(address=('localhost', 8080)):
    httpd = ThreadingServer(address, MothHandler)
    print("=== Listening on http://{}:{}/".format(address[0], address[1]))
    httpd.serve_forever()

if __name__ == '__main__':
    run()
