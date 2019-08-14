#!/usr/bin/python3

"""A validator for MOTH puzzles"""

import logging
import os
import os.path
import re

import moth

# pylint: disable=len-as-condition, line-too-long

DEFAULT_REQUIRED_FIELDS = ["answers", "authors", "summary"]

LOGGER = logging.getLogger(__name__)


class MothValidationError(Exception):

    """An exception for encapsulating MOTH puzzle validation errors"""


class MothValidator:

    """A class which validates MOTH categories"""

    def __init__(self, fields):
        self.required_fields = fields
        self.results = {"category": {}, "checks": []}

    def validate(self, categorydir, only_errors=False):
        """Run validation checks against a category"""
        LOGGER.debug("Loading category from %s", categorydir)
        try:
            category = moth.Category(categorydir, 0)
        except NotADirectoryError:
            return

        LOGGER.debug("Found %d puzzles in %s", len(category.pointvals()), categorydir)

        self.results["category"][categorydir] = {
            "puzzles": {},
            "name": os.path.basename(categorydir.strip(os.sep)),
        }
        curr_category = self.results["category"][categorydir]

        for check_function_name in [x for x in dir(self) if x.startswith("check_") and callable(getattr(self, x))]:
            if check_function_name not in self.results["checks"]:
                self.results["checks"].append(check_function_name)

        for puzzle in category:
            LOGGER.info("Processing %s: %s", categorydir, puzzle.points)

            curr_category["puzzles"][puzzle.points] = {}
            curr_puzzle = curr_category["puzzles"][puzzle.points]
            curr_puzzle["failures"] = []

            for check_function_name in [x for x in dir(self) if x.startswith("check_") and callable(getattr(self, x))]:
                check_function = getattr(self, check_function_name)
                LOGGER.debug("Running %s on %d", check_function_name, puzzle.points)

                try:
                    check_function(puzzle)
                except MothValidationError as ex:
                    curr_puzzle["failures"].append(str(ex))

            if only_errors and len(curr_puzzle["failures"]) == 0:
                del curr_category["puzzles"][puzzle.points]

    def check_fields(self, puzzle):
        """Check if the puzzle has the requested fields"""
        for field in self.required_fields:
            if not hasattr(puzzle, field):
                raise MothValidationError("Missing field %s" % (field,))

    @staticmethod
    def check_has_answers(puzzle):
        """Check if the puzle has answers defined"""
        if len(puzzle.answers) == 0:
            raise MothValidationError("No answers provided")

    @staticmethod
    def check_unique_answers(puzzle):
        """Check if puzzle answers are unique"""
        known_answers = []
        duplicate_answers = []

        for answer in puzzle.answers:
            if answer not in known_answers:
                known_answers.append(answer)
            else:
                duplicate_answers.append(answer)

        if len(duplicate_answers) > 0:
            raise MothValidationError("Duplicate answer(s) %s" % ", ".join(duplicate_answers))

    @staticmethod
    def check_has_authors(puzzle):
        """Check if the puzzle has authors defined"""
        if len(puzzle.authors) == 0:
            raise MothValidationError("No authors provided")

    @staticmethod
    def check_unique_authors(puzzle):
        """Check if puzzle authors are unique"""
        known_authors = []
        duplicate_authors = []

        for author in puzzle.authors:
            if author not in known_authors:
                known_authors.append(author)
            else:
                duplicate_authors.append(author)

        if len(duplicate_authors) > 0:
            raise MothValidationError("Duplicate author(s) %s" % ", ".join(duplicate_authors))

    @staticmethod
    def check_has_summary(puzzle):
        """Check if the puzzle has a summary"""
        if puzzle.summary is None:
            raise MothValidationError("Summary has not been provided")

    @staticmethod
    def check_has_body(puzzle):
        """Check if the puzzle has a body defined"""
        old_pos = puzzle.body.tell()
        puzzle.body.seek(0)
        if len(puzzle.body.read()) == 0:
            puzzle.body.seek(old_pos)
            raise MothValidationError("No body provided")

        puzzle.body.seek(old_pos)

    @staticmethod
    def check_ksa_format(puzzle):
        """Check if KSAs are properly formatted"""

        ksa_re = re.compile("^[KSA]\d{4}$")
        
        if hasattr(puzzle, "ksa"):
            for ksa in puzzle.ksa:
                if ksa_re.match(ksa) is None:
                    raise MothValidationError("Unrecognized KSA format (%s)" % (ksa,))


def output_json(data):
    """Output results in JSON format"""
    import json
    print(json.dumps(data))


def output_text(data):
    """Output results in a text-based tabular format"""

    longest_category = max([len(y["name"]) for x, y in data["category"].items()])
    longest_category = max([longest_category, len("Category")])
    longest_failure = len("Failures")
    for category_data in data["category"].values():
        for points, puzzle_data in category_data["puzzles"].items():
            longest_failure = max([longest_failure, len(", ".join(puzzle_data["failures"]))])

    formatstr = "| %%%ds | %%6s | %%%ds |" % (longest_category, longest_failure)
    headerfmt = formatstr % ("Category", "Points", "Failures")

    print(headerfmt)
    for cat_data in data["category"].values():
        for points, puzzle_data in sorted(cat_data["puzzles"].items()):
            print(formatstr % (cat_data["name"], points, ", ".join([str(x) for x in puzzle_data["failures"]])))


def main():
    """Main function"""
    # pylint: disable=invalid-name
    import argparse

    LOGGER.addHandler(logging.StreamHandler())

    parser = argparse.ArgumentParser(description="Validate MOTH puzzle field compliance")
    parser.add_argument("category", nargs="+", help="Categories to validate")
    parser.add_argument("-f", "--fields", help="Comma-separated list of fields to check for", default=",".join(DEFAULT_REQUIRED_FIELDS))

    parser.add_argument("-o", "--output-format", choices=["text", "json"], default="text", help="Output format (default: text)")
    parser.add_argument("-e", "--only-errors", action="store_true", default=False, help="Only output errors")
    parser.add_argument("-v", "--verbose", action="count", default=0, help="Increase verbosity of output, repeat to increase")

    args = parser.parse_args()

    if args.verbose == 1:
        LOGGER.setLevel("INFO")
    elif args.verbose > 1:
        LOGGER.setLevel("DEBUG")

    LOGGER.debug(args)
    validator = MothValidator(args.fields.split(","))

    for category in args.category:
        LOGGER.info("Validating %s", category)
        validator.validate(category, only_errors=args.only_errors)

    if args.output_format == "text":
        output_text(validator.results)
    elif args.output_format == "json":
        output_json(validator.results)


if __name__ == "__main__":
    main()
