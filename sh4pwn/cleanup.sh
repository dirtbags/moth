#! /bin/sh

# Invoked by buildroot just before making a filesystem

move () {
    [ -d $1 ] && (cd $1 && tar cf - .) | (cd $2 && tar xf -) && rm -rf $1
}

rm -rf $1/init.d
rm -rf $1/var/cache $1/var/lib

move $1/usr/bin $1/bin
move $1/usr/sbin $1/sbin
move $1/usr/lib $1/lib
move $1/usr/share $1/lib
[ -d $1/usr ] && rmdir $1/usr
[ -x $1/usr ] && rm $1/usr

set
cp $(dirname $0)/skeleton/sbin/init $1/sbin || exit 1

cat <<EOF > $1/etc/issue
 o     Dirtbags Shitty Linux
(m)         $(date --rfc-3339=date)
EOF

