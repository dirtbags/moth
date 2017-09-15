#!/bin/sh
#
# Script to clone and start a development server

set -e

if [ -f tools/devel-server.py ]; then
	cat <<EOM
This script is intended to be used to bootstrap a moth development server. It
looks like you're running the script from a moth repository working directory.

    $ mkdir /tmp/moth
    $ cd /tmp/moth
    $ curl https://raw.githubusercontent.com/dirtbags/moth/master/devel.sh | bash
EOM
	exit 1
fi

[ -d puzzles  ] || mkdir -p puzzles
[ -d moth/bin ] || git clone https://github.com/dirtbags/moth.git

cd moth
puzzles="$(readlink -e ../puzzles)"
ln -sf "${puzzles}" puzzles

printf "\n[+] Place puzzles at ${puzzles} ...\n"
python3 tools/devel-server.py
