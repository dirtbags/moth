PACKAGES = 

all: packages

define COPYTREE
	mkdir -p $(2)
	(cd $(1) && find . -not -name "*~" | cpio -o) | (cd $(2) && cpio -i)
endef


include */*.mk

packages: $(addsuffix .pkg, $(PACKAGES))

install: $(addsuffix -install, $(PACKAGES))

clean: $(addsuffix -clean, $(PACKAGES))
	rm -rf build *.pkg *-install *-build

%.pkg: %-install
	mksquashfs build/$* $*.pkg -all-root -noappend
