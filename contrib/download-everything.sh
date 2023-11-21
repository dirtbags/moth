#! /bin/sh

url=${1%/}
teamid=$2

case "$url:$teamid" in
    *:|-h*|--h*)
        cat <<EOD; exit 1
Usage: $0 MOTHURL TEAMID

Downloads all content currently open,
and writes it out to a zip file.

MOTHURL   URL to the instance
TEAMID    Team ID you used to log in
EOD
        ;;
esac

tmpdir=$(mktemp -d moth-dl.XXXXXX)
bye () {
    echo "bye now"
    rm -rf $tmpdir
}
trap bye EXIT

fetch () {
    curl -s -d id=$teamid "$@"
}

echo "=== Fetching puzzles and attachments"
fetch $url/state > $tmpdir/state.json
cat $tmpdir/state.json \
| jq -r '.Puzzles | to_entries[] | .key as $k | .value[] | select (. > 0) | "\($k)  \(.)"' \
| while read cat points; do
    echo "   + $cat $points"
    dir=$tmpdir/$cat/$points
    mkdir -p $dir
    fetch $url/content/$cat/$points/puzzle.json > $dir/puzzle.json
    cat $dir/puzzle.json | jq .Body > $dir/puzzle.html
    cat $dir/puzzle.json | jq -r '.Attachments[]?' | while read attachment; do
        echo "     - $attachment"
        fetch $url/content/$cat/$points/$attachment > $dir/$attachment
    done
done

zipfile=$(echo $url | grep -o '[a-z]*\.[a-z.]*').zip
echo "=== Writing $zipfile"
(cd $tmpdir && zip -r - .) > $zipfile

echo "=== Wrote $zipfile"
