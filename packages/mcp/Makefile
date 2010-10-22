SUBDIRS = src
PKGDIR = pkg
PACKAGE = mcp.pkg

all: package

package: $(PACKAGE)

$(PACKAGE): build
	mkdir -p $(PKGDIR)

	cp setup $(PKGDIR)

	find bin -not -name '*~' | cpio -p $(PKGDIR)
	cp src/in.tokend $(PKGDIR)/bin
	cp src/tokencli src/pointscli $(PKGDIR)/bin
	cp src/puzzles.cgi $(PKGDIR)/bin

	find service -not -name '*~' -not -name '#*' | cpio -p $(PKGDIR)

	find www -not -name '*~' -not -name '#*' | cpio -p $(PKGDIR)
	cp src/puzzler.cgi src/claim.cgi $(PKGDIR)/www

	mksquashfs $(PKGDIR) $(PACKAGE) -all-root -noappend


include $(addsuffix /*.mk, $(SUBDIRS))

test: build
	./test.sh

build: $(addsuffix -build, $(SUBDIRS))
clean: $(addsuffix -clean, $(SUBDIRS))
	rm -rf $(PKGDIR) $(PACKAGE)
