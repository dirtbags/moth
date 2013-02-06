#! /bin/sh -e

param () {
	ret=$(echo "$QUERY_STRING" | tr '=&' ' \n' | awk -v "k=$1" '($1==k) {print $2;}')
	ret=$(busybox httpd -d "$ret" || echo "$ret")
	echo "$ret"
}

team=$(param n)
hash=$(param h | tr -dc 0-9a-f)

cat <<EOF
Content-type: text/html

<!DOCTYPE html>
<html>
  <head>
    <title>Team Registration</title>
    <link rel="stylesheet" href="ctf.css" type="text/css">
  </head>
  <body>
    <h1>Team Registration</h1>
EOF

if [ -z "$hash" ]; then
	echo "<p>Empty hash, cannot complete request</p>"
elif ! grep -q $hash $CTF_BASE/state/teams/assigned.txt; then
	echo "<p>That hash has not been assigned.</p>"
elif [ -f $CTF_BASE/state/teams/names/$hash ]; then
	echo "<p>That hash has already been registered.</p>"
else
	printf "%s" "$team" > $CTF_BASE/state/teams/names/$hash
	echo "<p>Okay, your team has been named and you may begin using your hash!</p>"
fi

cat <<EOF
  </body>
</html>
EOF
