PACKAGES = 

all: packages

define COPYTREE
	mkdir -p $(2)
	(cd $(1) && find . -not -name "*~" | cpio -o) | (cd $(2) && cpio -i)
endef


include */*.mk

packages: $(addsuffix -package, $(PACKAGES))

clean: $(addsuffix -clean, $(PACKAGES))
	rm -rf build *.pkg
