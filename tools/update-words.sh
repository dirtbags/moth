#!/bin/sh
set +e

url='https://rawgit.com/first20hours/google-10000-english/master/google-10000-english-no-swears.txt'
getter="curl -sL"
fn="answer_words.txt"

filterer() {
	grep '......*'
}

if ! curl -h >/dev/null 2>/dev/null; then
	getter="wget -q -O -"
elif ! wget -h >/dev/null 2>/dev/null; then
	echo "[!] I don't know how to download. I need curl or wget."
fi

$getter "${url}" | filterer > ${fn}.tmp \
	&& mv -f ${fn}.tmp ${fn}
