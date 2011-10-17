#! /bin/sh

for i in $(seq 100); do
    if [ $(expr $i % 15) = 0 ]; then
        echo 'FizzBuzz'
    elif [ $(expr $i % 3) = 0 ]; then
        echo 'Fizz'
    elif [ $(expr $i % 5) = 0 ]; then
        echo 'Buzz'
    else
        echo $i
    fi
done
