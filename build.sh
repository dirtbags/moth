#! /bin/sh

set -e

version=$(date +%Y%m%d%H%M)

for img in moth moth-devel; do
    echo "==== $img"
    docker build --build-arg http_proxy=$http_proxy --tag dirtbags/$img --tag dirtbags/$img:$version -f Dockerfile.$img .
    [ "$1" = "--push" ] && docker push dirtbags/$img:$version && docker push dirtbags/$img
done
