from setuptools import setup

import sys
if sys.version_info < (3,5):
    sys.exit("Sorry, Python < 3.5 is not supported")

setup(
    name = "MOTHDevel",
    version = open("../VERSION", "r").read(),
    description = "The MOTH development toolkit",
    packages = ["MOTHDevel"],
    python_requires='~=3.5',
    include_package_data=True,
    entry_points = {
        "console_scripts": [
            "devel-server = MOTHDevel.devel_server:main",
        ],
    },
)
