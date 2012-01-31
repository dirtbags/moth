COWBULL_PKGDIR = $(TARGET)/cowbull

cowbull-install: cowbull-build
	mkdir -p $(COWBULL_PKGDIR)

	mkdir -p $(COWBULL_PKGDIR)/bin/
	$(MAKE) -C packages/cowbull/src install DESTDIR=$(CURDIR)/$(COWBULL_PKGDIR)

	$(call COPYTREE, packages/cowbull/service, $(COWBULL_PKGDIR)/service)
	cp packages/cowbull/tokens.txt $(COWBULL_PKGDIR)/

cowbull-clean:
	rm -rf $(COWBULL_PKGDIR)
	$(MAKE) -C packages/cowbull/src clean

cowbull-build:
	$(MAKE) -C packages/cowbull/src build

PACKAGES += cowbull
