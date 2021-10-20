#! /bin/sh

set -e

images="ghcr.io/dirtbags/moth"

ACTION=$1
if [ -z "$ACTION" ]; then
    echo "Usage: $0 ACTION"
    exit 1
fi

log () {
    printf "=== %s\n" "$*" 1>&2
}

fail () {
    printf "\033[31;1m=== FAIL: %s\033[0m\n" "$*" 1>&2
    exit 1
}

run () {
    printf "\033[32m\$\033[0m %s\n" "$*" 1>&2
    "$@"
}

tags () {
    pfx=$1
    for base in $images; do
        echo $pfx $base:${CI_COMMIT_REF_SLUG}
        echo $pfx $base:${CI_COMMIT_REF_SLUG%.*}
        echo $pfx $base:${CI_COMMIT_REF_SLUG%.*.*}
    done | uniq
}

case $ACTION in
    publish)
        run docker build \
            --file build/package/Containerfile \
            $(tags --tag) \
            .
        tags | while read image; do
            run docker push $image
        done
    ;;
*)
    echo "Unknown action: $1" 1>&2
    exit 1
    ;;
esac

