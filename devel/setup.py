from setuptools import setup

setup(
    name = "MOTHDevel",
    version = open("../VERSION", "r").read(),
    description = "The MOTH development toolkit",
    packages = ["MOTHDevel"],
    include_package_data=True,
    entry_points = {
        "console_scripts": [
            "devel-server = MOTHDevel.devel_server:main",
        ],
    },
)
