#!/bin/bash

workingDirectory=$1

if [ -z "${workingDirectory}" ] || [ "$1" == "-clean" ]; then
	workingDirectory=$PWD
fi

if [ "$1" == "-clean" ] || [ "$2" == "-clean" ]; then
	echo "Cleaning"
	rm -rf "$workingDirectory/compiled-puzzles"
	rm -rf "$workingDirectory/mothballs"
	rm -rf "$workingDirectory/state"
	exit
fi

echo "Working in directory $workingDirectory"

bash $workingDirectory/build.sh

mkdir $workingDirectory/compiled-puzzles

FILES=$workingDirectory/example-puzzles/*
for f in $FILES
do
	$workingDirectory/devel/package-puzzles.py $workingDirectory/compiled-puzzles $f
done

mkdir $workingDirectory/mothballs
cp $workingDirectory/compiled-puzzles/* $workingDirectory/mothballs

mkdir $workingDirectory/state

sudo docker run --rm -it -p 8080:8080 -v $workingDirectory/state:/state -v $workingDirectory/mothballs:/mothballs dirtbags/moth
