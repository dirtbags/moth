#! /bin/sh -e

DESTDIR="$1"

if [ -z "$DESTDIR" ]; then
	echo "Usage: $0 DESTDIR"
	exit
fi

cd $(dirname $0)

older () {
	[ -z "$1" ] && return 1
	target="$1"; shift
	[ -f "$target" ] || return 0
	for i in "$@"; do
		[ "$i" -nt "$target" ] && return 0
	done
	return 1
}

copy () {
	src="$1"
	target="$2/$src"
	targetdir=$(dirname "$target")
	if older "$target" "$src"; then
		echo "COPY $src"
		mkdir -p "$targetdir"
		cp "$src" "$target"
	fi
}

setup() {
	www="$1"
	[ -d "$DESTDIR/state" ] && return
	echo "SETUP"
	for i in points.new points.tmp teams; do
		dir="$DESTDIR/state/$i"
		mkdir -p "$dir"
		setfacl -m ${www}:rwx "$dir"
	done
	mkdir -p "$DESTDIR/packages"
	>> "$DESTDIR/state/points.log"
	if ! [ -f "$DESTDIR/assigned.txt" ]; then
		hd </dev/urandom | awk '{print $3 $4 $5 $6;}' | head -n 100 > "$DESTDIR/assigned.txt"
	fi

	mkdir -p "$DESTDIR/www"
	ln -sf ../state/points.json "$DESTDIR/www"
	ln -sf ../state/puzzles.json "$DESTDIR/www"
}


echo "Figuring out web user..."
for www in www-data http tc _ _www; do
	id $www && break
done
if [ $www = _ ]; then
	echo "Unable to determine httpd user on this system. Dying."
	exit 1
fi

mkdir -p "$DESTDIR" || exit 1

setup $www
git $SRCDIR ls-files | while read fn; do
case "$fn" in
	example-puzzles/*|tools/*|docs/*|install.sh|setup.cfg|README.md|.gitignore|src/mothd)
		true # skip
		;;
	www/*)
		copy "$fn" "$DESTDIR/"
		;;
	bin/*)
		copy "$fn" "$DESTDIR/"
		;;
	*)
		echo "??? $fn"
		;;
esac
done

echo "All done installing."
