MULTICASTER_PKGDIR = $(TARGET)/multicaster

multicaster-install: multicaster-build
	mkdir -p $(MULTICASTER_PKGDIR)

	mkdir -p $(MULTICASTER_PKGDIR)/bin/
	$(MAKE) -C packages/multicaster/src install DESTDIR=$(CURDIR)/$(MULTICASTER_PKGDIR)

multicaster-clean:
	rm -rf $(MULTICASTER_PKGDIR)
	$(MAKE) -C packages/multicaster/src clean

multicaster-build:
	$(MAKE) -C packages/multicaster/src build

PACKAGES += multicaster
