#! /bin/sh -e

TYPE=p2

case ${1:-$TYPE} in
    mcp)
        packages='mcp net-re'
        ;;
    router)
        packages='router'
        ;;
    p2)
        packages='p2 gs archaeology nocode steg js proto'
        ;;
    p2cli)
        packages='p2client'
        ;;
esac

PATH=$HOME/src/buildroot/output/host/usr/bin:$PATH
for arch in arm i386; do
    command -v ${arch}-linux-cc && ARCH=${arch}-linux export ARCH
done

if [ -z "$ARCH" ]; then
    echo "I can't find a cross-compiler."
    exit 1
fi

make -C $HOME/src/puzzles
make -C $HOME/src/ctf

for p in $packages; do
    for pd in ctf puzzles; do
        pp=$HOME/src/$pd/bin/$p.pkg
        [ -f $pp ] && op="$op $pp"
    done
done

mksquashfs \
    $op \
    $HOME/ctf.squashfs -noappend

echo $ARCH

if [ $ARCH = i386-linux ]; then
    lsmod | grep -q kvm-intel || sudo modprobe kvm-intel
    sudo qemu-system-i386 \
        -nographic \
        -kernel $HOME/src/buildroot/output/images/bzImage \
        -initrd $HOME/src/buildroot/output/images/rootfs.squashfs \
        -append "console=ttyS0 packages=/dev/sda ipv6 debug" \
        -hda $HOME/ctf.squashfs \
        -net nic,model=e1000 \
        -net tap,vlan=0,script=$HOME/src/ctf/qemu-ifup,downscript=/bin/true
fi
