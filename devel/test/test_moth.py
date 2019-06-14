import io
import os.path
import tempfile
import unittest

from MOTHDevel import moth


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

