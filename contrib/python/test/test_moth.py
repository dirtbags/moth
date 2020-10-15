import io
import os.path
import tempfile
import unittest

import moth

class TestMoth(unittest.TestCase):

    def test_djb2hash(self):
        input_data = "test"
        expected = 2090756197
        self.assertEqual(moth.djb2hash(input_data), expected)

    def test_log(self):
        puzzle = moth.Puzzle(12345, 1)
        message = "Test message"
        puzzle.log(message)
        self.assertIn(message, puzzle.logs)

    def test_random_hash(self):
        puzzle = moth.Puzzle(12345, 1)
        self.assertEqual(len(puzzle.random_hash()), 8)

    def test_random_hash_repeatable(self):
        puzzle1 = moth.Puzzle(12345, 1)
        puzzle2 = moth.Puzzle(12345, 1)
        puzzle3 = moth.Puzzle(11111, 1)
        puzzle4 = moth.Puzzle(12345, 2)

        p1_hash = puzzle1.random_hash()
        p2_hash = puzzle2.random_hash()
        p3_hash = puzzle3.random_hash()
        p4_hash = puzzle4.random_hash()

        self.assertEqual(p1_hash, p2_hash)
        self.assertNotEqual(p1_hash, p3_hash)
        self.assertNotEqual(p1_hash, p4_hash)

    def test_make_temp_file(self):
        puzzle = moth.Puzzle(12345, 1)
        tt = puzzle.make_temp_file(name="Test stream")
        tt.write(b"Test")
        self.assertIn("Test stream", puzzle.files)
        tt.seek(0)
        self.assertEqual(puzzle.files["Test stream"].stream.read(), b"Test")

    def test_add_stream_visible(self):
        puzzle = moth.Puzzle(12345, 1)
        data = b"Test"
        with io.BytesIO(data) as buf:
            puzzle.add_stream(buf, name="Test stream", visible=True)
            self.assertIn("Test stream", puzzle.files)
            self.assertEqual(puzzle.files["Test stream"].stream.read(), data)
            self.assertEqual(puzzle.files["Test stream"].visible, True)

    def test_add_stream_notvisible(self):
        puzzle = moth.Puzzle(12345, 1)
        data = b"Test"

        with io.BytesIO(data) as buf:
            puzzle.add_stream(buf, name="Test stream", visible=False)
            self.assertIn("Test stream", puzzle.files)
            self.assertEqual(puzzle.files["Test stream"].stream.read(), data)
            self.assertEqual(puzzle.files["Test stream"].visible, False)

    def test_add_stream_visible_no_name(self):
        puzzle = moth.Puzzle(12345, 1)
        data = b"Test"

        with io.BytesIO(data) as buf:
            puzzle.add_stream(buf, visible=True)
            self.assertGreater(len(puzzle.files), 0)

    def test_add_file(self):
        puzzle = moth.Puzzle(12345, 1)
        data = b"Test"
        with tempfile.NamedTemporaryFile() as tf:
            tf.write(data)
            tf.flush()
            puzzle.add_file(tf.name)
            self.assertIn(os.path.basename(tf.name), puzzle.files)
            self.assertEqual(puzzle.files[os.path.basename(tf.name)].stream.read(), data)
            self.assertEqual(puzzle.files[os.path.basename(tf.name)].visible, True)

    def test_invalid_header(self):
        puzzle = moth.Puzzle(12345, 1)
        stream = """author: foo
                 answer: 10
                 baz: stuff


                 Some body
                 """

        with io.StringIO(stream) as buff:
            with self.assertRaisesRegex(ValueError, "Unrecognized header field: baz"):
                puzzle.read_stream(buff)

    def test_known_headers_moth(self):
        puzzle = moth.Puzzle(12345, 1)
        stream = """author: foo
                 answer: the answer is answer
                 summary: read the puzzle
                 pattern: a matching pattern
                 hint: This is a helpful hint
                 name: No idea what this is for

                 Some body
                 """

        with io.StringIO(stream) as buff:
            puzzle.read_stream(buff)
            pkg = puzzle.package()

    def test_hexdump(self):
        puzzle = moth.Puzzle(12345, 1)
        test_data = [0, 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17]
        puzzle.hexdump(test_data)

    def test_hexump_none(self):
        puzzle = moth.Puzzle(12345, 1)
        test_data = [0, 1, 2, 3, None, 5, 6, 7, 8]
        puzzle.hexdump(test_data)

    def test_hexdump_elided_dupe_row(self):
        puzzle = moth.Puzzle(12345, 1)
        test_data = [1 for x in range(4*16)]
        puzzle.hexdump(test_data)

    def test_category(self):
        category = moth.Category("./test/category_test", 1)
        self.assertIsNotNone(category.catmod)

        puzzles = list(category)

    def test_authors_legacy(self):
        puzzle = moth.Puzzle(12345, 1)
        puzzle.author = "foo"

        self.assertEqual(puzzle.get_authors(), ["foo"])

    def test_authors(self):
        puzzle = moth.Puzzle(12345, 1)

        self.assertEqual(puzzle.get_authors(), [])
