#! /bin/sh

echo 'Content-type: application/octet-stream'
echo

tar czf - /var/lib/ctf | KEY=crashmaster arc4
