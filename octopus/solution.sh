#! /bin/sh

port=8888

blooper=$(tempfile)
trap "rm $blooper" 0

echo foo | socat -t 0.01 STDIO UDP:127.0.0.1:8888 | tail -n +4 > $blooper

for i in $(seq 8); do
    result=$(socat -t 0.01 STDIO UDP:127.0.0.1:$port < $blooper | awk -F': ' '(NF > 1) {print $2; exit;}')
    port=$(echo "ibase=8; $result" | bc)
    echo $port
done
echo $result