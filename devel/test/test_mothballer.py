import os
import os.path
import unittest
import tempfile
import zipfile

from MOTHDevel import mothballer


class TestMothballer(unittest.TestCase):

    test_categories = ["counting", "example"]

    def test_main(self):
        with tempfile.TemporaryDirectory() as td:
            args = [td]
            args.extend([os.path.join("..", "example-puzzles", x) for x in self.test_categories])
            mothballer.main(args)

    def test_write_kv_pair_1(self):
        data = {"key1": "value1", "key2": "value2", "listkey1": ["1", "2", "3"]}

        with tempfile.TemporaryFile() as tf:
            with zipfile.ZipFile(tf, mode="w") as zf:
                mothballer.write_kv_pairs(zf, "testkeys", data)
                self.assertEqual(1, len(zf.namelist()))

    def test_escape_1(self):
        in_str = "Test string &"
        self.assertEqual(mothballer.escape(in_str), "Test string &amp;")

    def test_escape_2(self):
        in_str = "Test string <"
        self.assertEqual(mothballer.escape(in_str), "Test string &lt;")

    def test_escape_3(self):
        in_str = "Test string >"
        self.assertEqual(mothballer.escape(in_str), "Test string &gt;")

    def test_compile_example_puzzles(self):
        with tempfile.TemporaryDirectory() as td:
            for category in self.test_categories:
                categorydir = os.path.join("..", "example-puzzles", category)
                mothballer.build_category(categorydir, td)

            for root, dirs, files in os.walk(td):
                self.assertEqual(len(files), len(self.test_categories))
                for filename in files:
                    with self.subTest(filename=filename):
                        self.assertIn(filename, ["%s.mb" % x for x in self.test_categories])
                        with zipfile.ZipFile(os.path.join(root, filename)) as zf:
                            zip_contents = zf.namelist()
                            self.assertIn("category_seed.txt", zip_contents)
                            self.assertIn("map.txt", zip_contents)
                            self.assertIn("answers.txt", zip_contents)
                            self.assertIn("summaries.txt", zip_contents)
                            for puzzle in zf.open("map.txt"):
                                points, subpath = puzzle.strip().split()
                                puzzle_json_path = os.path.join(b"content", subpath, b"puzzle.json")
                                self.assertIn(puzzle_json_path.decode("ascii"), zip_contents)


    def test_compile_example_puzzles_rebuild(self):
        
        """Test cases where SEED is defined in an existing mothball"""

        with tempfile.TemporaryDirectory() as td:
            for category in self.test_categories:
                with self.subTest(category=category):
                    categorydir = os.path.join("..", "example-puzzles", category)
                    mothballer.build_category(categorydir, td)

                    seed = None
                    new_seed = None

                    with zipfile.ZipFile(os.path.join(td, "%s.mb" % (category,)), mode="a") as zf:
                        seed = zf.open("category_seed.txt").read()
                        zf.writestr(mothballer.SEEDFN, seed)

                    self.assertIsNotNone(seed)

                    mothballer.build_category(categorydir, td)
                    with zipfile.ZipFile(os.path.join(td, "%s.mb" % (category,))) as zf:
                        new_seed = zf.open("category_seed.txt").read()

                    self.assertEqual(seed, new_seed)
