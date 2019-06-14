from setuptools import setup

import sys
if sys.version_info < (3,5):
    sys.exit("Sorry, Python < 3.5 is not supported")

setup(
    name = "MOTHDevel",
    version = open("../VERSION", "r").read().strip(),
    description = "The MOTH development toolkit",
    packages = ["MOTHDevel"],
    python_requires='~=3.5',
    tests_require=["flake8==3.7.7", "nose>=1.3.7", "pylint==2.3.1", "coverage==4.5.3"],
    extras_require={
        "scapy": ["scapy>=2.4.2"],
    },
    include_package_data=True,
    entry_points = {
        "console_scripts": [
            "devel-server = MOTHDevel.devel_server:main",
            "mothballer = MOTHDevel.mothballer:main",
        ],
    },
)
