import os
import os.path
import pathlib
import requests
import socket
from threading import Thread
import unittest

from MOTHDevel.devel_server import MothServer, MothRequestHandler

def get_free_port():
    s = socket.socket(socket.AF_INET, type=socket.SOCK_STREAM)
    s.bind(("localhost", 0))
    address, port = s.getsockname()
    s.close()
    return port


class TestDevelServer(unittest.TestCase):

    @classmethod
    def setup_class(cls):
        cls.mock_server_port = get_free_port()
        cls.mock_server = MothServer(("0.0.0.0", cls.mock_server_port), MothRequestHandler)
        cls.mock_server_thread = Thread(target=cls.mock_server.serve_forever)
        cls.mock_server_thread.setDaemon(True)
        cls.mock_server.args["puzzles_dir"] = pathlib.Path("../example-puzzles")
        cls.mock_server.args["theme_dir"] = "../theme"
        cls.mock_server_thread.start()

    def test_head(self):
        url = "http://localhost:%d/" % (self.mock_server_port,)
        res = requests.head(url)
        self.assertEqual(res.status_code, 501)

    def test_index(self):
        url = "http://localhost:%d" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "text/html; charset=utf-8")

    def test_puzzle_list(self):
        url = "http://localhost:%d/1/puzzles.json" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")

    def test_puzzle_list_page(self):
        url = "http://localhost:%d/1/" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "text/html")

    def test_puzzle(self):
        url = "http://localhost:%d/1/content/counting/1/puzzle.json" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")

    def test_answer(self):
        url = "http://localhost:%d/1/answer" % (self.mock_server_port,)
        res = requests.post(url, data={"cat": "counting", "points": 1, "answer": 9})
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")
        self.assertEqual(res.json()["data"]["description"], "Answer is correct")
        self.assertEqual(res.json()["data"]["short"], "")
        self.assertEqual(res.json()["status"], "success")

    def test_incorrect_answer(self):
        url = "http://localhost:%d/1/answer" % (self.mock_server_port,)
        res = requests.post(url, data={"cat": "counting", "points": 1, "answer": 1})
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")
        self.assertEqual(res.json()["data"]["description"], "Provided answer was not in list of answers")
        self.assertEqual(res.json()["data"]["short"], "")
        self.assertEqual(res.json()["status"], "success")

    @unittest.skip("GETs are not yet supported")
    def test_answer(self):
        url = "http://localhost:%d/1/answer" % (self.mock_server_port,)
        res = requests.get(url, params={"cat": "counting", "points": 1, "answer": 9})
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")
        self.assertEqual(res.json()["data"]["description"], "Answer is correct")
        self.assertEqual(res.json()["data"]["short"], "")
        self.assertEqual(res.json()["status"], "success")

    @unittest.skip("GETs are not yet supported")
    def test_incorrect_answer(self):
        url = "http://localhost:%d/1/answer" % (self.mock_server_port,)
        res = requests.get(url, params={"cat": "counting", "points": 1, "answer": 1})
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")
        self.assertEqual(res.json()["data"]["description"], "Provided answer was not in list of answers")
        self.assertEqual(res.json()["data"]["short"], "")
        self.assertEqual(res.json()["status"], "success")

    def test_get_puzzlefile(self):
        url = "http://localhost:%d/1/content/example/2/s.jpg" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "image/jpeg")

    def test_get_unknown_puzzlefile(self):
        url = "http://localhost:%d/1/content/example/2/invalid.jpg" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 404)
        self.assertEqual(res.headers["Content-Type"], "text/html;charset=utf-8")

    def test_mothballer(self):
        url = "http://localhost:%d/1/mothballer/example.mb" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/octet_stream")

    @unittest.skip("Mothballer currently returns empty mothballs for invalid categories")
    def test_mothballer_invalid_category(self):
        url = "http://localhost:%d/1/mothballer/invalid.mb" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "text/html; charset=utf-8")

    def test_static_resource(self):
        url = "http://localhost:%d/1/index.html" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "text/html")


class TestDevelServerMisconfigured(unittest.TestCase):

    @classmethod
    def setup_class(cls):
        cls.mock_server_port = get_free_port()
        cls.mock_server = MothServer(("0.0.0.0", cls.mock_server_port), MothRequestHandler)
        cls.mock_server_thread = Thread(target=cls.mock_server.serve_forever)
        cls.mock_server_thread.setDaemon(True)
        cls.mock_server.args["puzzles_dir"] = pathlib.Path("../invalid/")
        cls.mock_server.args["theme_dir"] = "../theme"
        cls.mock_server_thread.start()

    def test_puzzle_list(self):
        url = "http://localhost:%d/1/puzzles.json" % (self.mock_server_port,)
        res = requests.get(url)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.headers["Content-Type"], "application/json")

