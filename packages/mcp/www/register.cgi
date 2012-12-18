#! /bin/sh -e

team=$(echo "$QUERY_STRING" | sed -n s'/.*team=\([^&]*\).*/\1/p')
team=$(busybox httpd -d "$team" || echo "$team")

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

if [ ! -w $CTF_BASE/www ] || [ ! -w $CTF_BASE/state/teams ]; then
    echo "<p>It looks like the server isn't set up for self-registrations."
    echo "Go talk to someone at the head table to register your team.</p>"
else
    echo "<p>Team name: $team</p>"
    echo -n "<pre>"
    if $CTF_BASE/mcp/bin/addteam "$team"; then
        echo "</pre><p>Write this hash down.  You will use it to claim points.</p>"
    else
        echo "Oops, something broke.  Better call Neale.</pre>"
    fi
fi
cat <<EOF
  </body>
</html>
EOF
