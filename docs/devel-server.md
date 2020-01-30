Using the MOTH Development Server
======================

To make puzzle development easier,
MOTH comes with a standalone web server written in Python,
which will show you how your puzzles are going to look without making you compile or package anything.

It even works in Windows,
because that is what my career has become.


Getting It Going
----------------

### With Docker

If you can use docker, you are in luck:

	docker run --rm -t -p 8080:8080 dirtbags/moth-devel

Gets you a development puzzle server running on port 8080,
with the sample puzzle directory set up.


### Without Docker

If you can't use docker,
try this:

	apt install python3
	pip3 install scapy pillow PyYAML
	git clone https://github.com/dirtbags/moth/
	cd moth
	python3 devel/devel-server.py --puzzles example-puzzles


Installing New Puzzles
-----------------------------

The development server wants to see category directories under `puzzles`,
like this:

	$ find puzzles -type d
	puzzles/
	puzzles/category1/
	puzzles/category1/10/
	puzzles/category1/20/
	puzzles/category1/30/
	puzzles/category2/
	puzzles/category2/100/
	puzzles/category2/200/
	puzzles/category2/300/


### With Docker

	docker run --rm -t -v /path/to/my/puzzles:/puzzles:ro -p 8080:8080 dirtbags/moth-devel


### Without Docker

You can use the `--puzzles` argument to `devel-server.py`
to specify a path to your puzzles directory.
