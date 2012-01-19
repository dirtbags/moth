#! /bin/sh

## Run like this:
##
##    socat EXEC:./solution.sh EXEC:./revwords 3<token.txt
##

lrev () {
    while [ -n "$1" ]; do
        echo $1 | rev
        shift
    done
}

while read line; do
    echo $line 1>&2
    enil=$(lrev $line)
    echo $enil
done