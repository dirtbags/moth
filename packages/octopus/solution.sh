#! /bin/sh

port=8888
host=${1:-[::1]}

blooper=$(tempfile)
trap "rm $blooper" 0

echo foo | socat -t 0.01 STDIO UDP6:$host:$port | tail -n +5 > $blooper

for i in $(seq 8); do
    result=$(socat -t 0.01 STDIO UDP6:$host:$port < $blooper | awk -F': ' '(NF > 1) {print $2; exit;}')
    port=$(printf "%d" "0$result")
    echo "next port: $port ($result)"
done
echo $result
