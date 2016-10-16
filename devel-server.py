#!/usr/bin/python3

import glob
import http.server
import mistune
import pathlib
import socketserver

HTTPStatus = http.server.HTTPStatus

def page(title, body):
    return """<!DOCTYPE html>
<html>
  <head>
    <title>{}</title>
    <link rel="stylesheet" href="/files/www/res/style.css">
  </head>
  <body>
    <div id="body" class="terminal">
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

class MothHandler(http.server.CGIHTTPRequestHandler):
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
        parts = self.path.split("/")
        if len(parts) < 3:
            body.append("# Puzzle Categories")
            # List all categories
            for i in glob.glob("puzzles/*/"):
                body.append("* [{}](/{})".format(i, i))
        else:
            body.append("# Not Implemented Yet")
        self.serve_md('\n'.join(body))

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

        self.send_response(http.server.HTTPStatus.OK)
        self.send_header("Content-type", "text/html; encoding=utf-8")
        self.send_header("Content-Length", len(content))
        try:
            fs = fspath.stat()
            self.send_header("Last-Modified", self.date_time_string(fs.st_mtime))
        except:
            pass
        self.end_headers()
        self.wfile.write(content.encode('utf-8'))

def run(address=('', 8080)):
    httpd = ThreadingServer(address, MothHandler)
    print("=== Listening on http://{}:{}/".format(address[0] or "localhost", address[1]))
    httpd.serve_forever()

if __name__ == '__main__':
    run()
