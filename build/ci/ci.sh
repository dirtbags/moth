#! /bin/sh

set -e

ACTION=$1
BASE=$2
if [ -z "$ACTION" ] || [ -z "$BASE" ]; then
    echo "Usage: $0 ACTION BASE"
    exit 1
fi

log () {
    printf "=== %s\n" "$*" 1>&2
}

fail () {
    printf "\033[31;1m=== FAIL: %s\033[0m\n" "$*" 1>&2
    exit 1
}

tags () {
    pfx=$1
    for base in ghcr.io/dirtbags/moth dirtbags/moth; do
        echo $pfx $base:${CI_COMMIT_REF_SLUG}
        echo $pfx $base:${CI_COMMIT_REF_SLUG%.*}
        echo $pfx $base:${CI_COMMIT_REF_SLUG%.*.*}
    done | uniq
}

case $ACTION in
    publish)
        docker build \
            --file build/package/Containerfile \
            $(tags)
        docker push $(tags --destination)
    ;;
*)
    echo "Unknown action: $1" 1>&2
    exit 1
    ;;
esac

