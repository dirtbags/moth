#! /bin/sh

set -e

cd $(dirname $0)
base=../..

VERSION=$(cat $base/CHANGELOG.md | awk -F '[][]' '/^## \[/ {print $2; exit}')
GO_VERSION=$(cat $base/go.mod | sed -n 's/^go //p')

(
    zipfile=winmoth.$VERSION.zip
    echo "=== Building $zipfile"
    mkdir -p winmoth winmoth/state winmoth/puzzles winmoth/mothballs
    echo devel > winmoth/state/teamids.txt
    cp moth-devel.bat winmoth
    cp -a $base/theme winmoth
    (
        cd winmoth
        GOOS=windows GOARCH=amd64 go build ../$base/cmd/mothd/...
    )
    zip -r $zipfile winmoth

    rm -rf winmoth
)

tag=dirtbags/moth:$VERSION
echo "==== Building $tag"
docker build \
    --build-arg GO_VERSION=$GO_VERSION \
    --build-arg http_proxy --build-arg https_proxy --build-arg no_proxy \
    --tag $tag \
    -f Containerfile $base

exit 0
