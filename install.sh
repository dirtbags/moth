#! /bin/sh -e

DRIVE=$1
if ! [ -b "$DRIVE" ]; then
    echo "Usage: $0 DEVICE"
    echo
    echo "Prepares DEVICE with Capture The Flag goodness"
    exit
fi

size=$(sfdisk -s $DRIVE)
fatsize=$(expr $size \* 95 / 100)

FATFS=${DRIVE}1
EXTFS=${DRIVE}2

sfdisk -uB $DRIVE <<EOF
,$fatsize,6,*
,,L
EOF
sync

mkdir -p /mnt/ctf-install

mkdosfs -n PACKAGES $FATFS
mke2fs -j -L VAR $EXTFS

cat mbr.bin > $DRIVE
mount $FATFS /mnt/ctf-install
mkdir /mnt/ctf-install/syslinux
umount /mnt/ctf-install
syslinux -d syslinux $FATFS

mount $FATFS /mnt/ctf-install
cp rootfs.squashfs /mnt/ctf-install/syslinux/dbtl.squashfs
cp bzImage /mnt/ctf-install/syslinux/

cat <<EOD >/mnt/ctf-install/syslinux/syslinux.cfg
DEFAULT ctf
LABEL ctf
  KERNEL bzImage
  INITRD dbtl.squashfs

LABEL dbtl
  KERNEL bzImage
  INITRD dbtl.squashfs
  APPEND packages=disabled
EOD

cp $(basename $0)/../bin/*.pkg /mnt/ctf-install
umount /mnt/ctf-install
rmdir /mnt/ctf-install

sync

echo "Done"
