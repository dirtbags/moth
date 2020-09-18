#! /bin/sh

set -e

cd $(dirname $0)/../..

PODMAN=$(command -v podman || echo docker)
VERSION=$(cat CHANGELOG.md | awk -F '[][]' '/^## \[/ {print $2; exit}')

for target in moth moth-devel; do
    tag=dirtbags/$target:$VERSION
    echo "==== Building $tag"
    $PODMAN build \
        --build-arg http_proxy --build-arg https_proxy --build-arg no_proxy \
        --tag dirtbags/$target \
        --target $target \
        -f build/package/Containerfile .
done

exit 0
