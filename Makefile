# Scratch directory for building extrenal sources
BUILD = build

# Root to install things before they're packaged
TARGET = target

# Downloaded source files go here
CACHE = cache

# The end result
BIN = bin

# Things configure likes to see
CONFIG_XCOMPILE_FLAGS = --host=i386-linux --program-transform-name=

all: packages

dist: ctf-install.zip
ctf-install.zip: packages.zip bzImage rootfs.squashfs /usr/lib/syslinux/mbr.bin
	zip --junk-paths $@ packages.zip bzImage rootfs.squashfs /usr/lib/syslinux/mbr.bin install.sh

packages.zip: packages
	zip --junk-paths $@ bin/*.pkg

clean: packages-clean
	rm -rf $(BUILD) $(TARGET) $(BIN)

scrub: clean
	rm -rf $(CACHE)

-include */*.mk
