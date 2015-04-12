#! /bin/sh -e

# Change to CTF_BASE
cd ${CTF_BASE:-.}
for i in $(seq 5); do
	[ -f assigned.txt ] && break
	cd ..
done
if ! [ -f assigned.txt ]; then
	cat <<EOF
Content-type: text/html

Cannot find CTF_BASE
EOF
	exit 1
fi

# Read CGI parameters
param () {
	ret=$(echo "$QUERY_STRING" | awk -F '=' -v 'RS=&' -v "k=$1" '($1==k) {print $2;}')
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
		<meta charset="UTF-8">
		<title>Team Registration</title>
		<link rel="stylesheet" href="css/style.css" type="text/css">
	</head>
	<body>
		<h1>Team Registration</h1>
		<section>
EOF

if [ -z "$hash" ] || [ -z "$team" ]; then
	echo "<h2>Oops!</h2>"
	echo "<p>Empty field, cannot complete request</p>"
elif ! grep -q "^$hash$" assigned.txt; then
	echo "<h2>Oops!</h2>"
	echo "<p>That hash has not been assigned.</p>"
elif [ -f state/teams/$hash ]; then
	echo "<h2>Oops!</h2>"
	echo "<p>That hash has already been registered.</p>"
else
	printf "%s" "$team" > state/teams/$hash
	echo "<h2>Success!</h2>"
	echo "<p>Okay, your team has been named and you may begin using your hash!</p>"
fi

cat <<EOF
		</section>
		<nav>
			<ul>
				<li><a href="register.html">Register</a></li>
				<li><a href="puzzles.html">Puzzles</a></li>
				<li><a href="scoreboard.html">Scoreboard</a></li>
			</ul>
		</nav>
	</body>
</html>
EOF
