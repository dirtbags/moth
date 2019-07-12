#!/usr/bin/python3

import logging

import moth

DEFAULT_REQUIRED_FIELDS = ["answers", "authors", "summary"]

LOGGER = logging.getLogger(__name__)


class MothValidationError(Exception):

    pass


class MothValidator:

    def __init__(self, fields):
        self.required_fields = fields
        self.results = {}

    def validate(self, categorydir):
        LOGGER.debug("Loading category from %s", categorydir)
        category = moth.Category(categorydir, 0)
        LOGGER.debug("Found %d puzzles in %s", len(category.pointvals()), categorydir)

        self.results[categorydir] = {}
        curr_category = self.results[categorydir]

        for puzzle in category:
            LOGGER.info("Processing %s: %s", categorydir, puzzle.points)

            curr_category[puzzle.points] = {}
            curr_puzzle = curr_category[puzzle.points]
            curr_puzzle["checks"] = []
            curr_puzzle["failures"] = []

            for check_function_name in [x for x in dir(self) if x.startswith("check_") and callable(getattr(self, x))]:
                check_function = getattr(self, check_function_name)
                LOGGER.debug("Running %s on %d", check_function_name, puzzle.points)

                curr_puzzle["checks"].append(check_function_name)

                try:
                    check_function(puzzle)
                except MothValidationError as ex:
                    curr_puzzle["failures"].append(ex)
                    LOGGER.exception(ex)


    def check_fields(self, puzzle):
        for field in self.required_fields:
            if not hasattr(puzzle, field):
                raise MothValidationError("Missing field %s" % (field,))

    def check_has_answers(self, puzzle):
        if len(puzzle.answers) == 0:
            raise MothValidationError("No answers provided")

    def check_has_authors(self, puzzle):
        if len(puzzle.authors) == 0:
            raise MothValidationError("No authors provided")

    def check_has_summary(self, puzzle):
        if puzzle.summary is None:
            raise MothValidationError("Summary has not been provided")

    def check_has_body(self, puzzle):
        old_pos = puzzle.body.tell()
        puzzle.body.seek(0)
        if len(puzzle.body.read()) == 0:
            puzzle.body.seek(old_pos)
            raise MothValidationError("No body provided")
        else:
            puzzle.body.seek(old_pos)

    # Leaving this as a placeholder until KSAs are formally supported
    def check_ksa_format(self, puzzle):
        if hasattr(puzzle, "ksa"):
            for ksa in puzzle.ksa:
                if not ksa.startswith("K"):
                    raise MothValidationError("Unrecognized KSA format")

def output_json(data):
    import json
    print(json.dumps(data))

def output_text(data):
    for category, cat_data in data.items():
        print("= %s =" % (category,))
        print("| Points | Checks | Errors |")
        for points, puzzle_data in cat_data.items():
            print("| %d | %s | %s |" % (points, ", ".join(puzzle_data["checks"]), puzzle_data["failures"]))
        


if __name__ == "__main__":
    import argparse

    LOGGER.addHandler(logging.StreamHandler())

    parser = argparse.ArgumentParser(description="Validate MOTH puzzle field compliance")
    parser.add_argument("category", nargs="+", help="Categories to validate")
    parser.add_argument("-f", "--fields", help="Comma-separated list of fields to check for", default=",".join(DEFAULT_REQUIRED_FIELDS))

    parser.add_argument("-o", "--output-format", choices=["text", "json", "csv"], default="text", help="Output format (default: text)")
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
        validator.validate(category)

    if args.output_format == "text":
        output_text(validator.results)
    elif args.output_format == "json":
        output_json(validator.results)
