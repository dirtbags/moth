#! /bin/sh

cd ${CTF_BASE:-/var/lib/ctf}/teams/names

escape () {
    sed 's/&/\&amp;/g;s/</\&lt;/g;s/>/\&gt;/g'
}

title='Teams'

cat <<EOF
<!DOCTYPE html>
<html>
  <head>
    <title>$title</title>
    <link rel="stylesheet" href="ctf.css" type="text/css">
  </head>
  <body>
    <h1>$title</h1>
EOF

echo "<table>"
echo "<tr><th>Team</th><th>ID</th></tr>"
for i in *; do
    echo "<tr><td>"
    escape < $i
    echo "</td><td><samp>$i</samp></td></tr>"
done
echo "</table>"

cat <<EOF
  <p>
    Team names are only used on this page and the scoreboard.
    Use your team ID to claim points.
  </p>
  </body>
</html>
EOF
