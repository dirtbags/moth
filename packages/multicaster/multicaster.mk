MULTICASTER_PKGDIR = $(TARGET)/multicaster

multicaster-install: multicaster-build
	mkdir -p $(MULTICASTER_PKGDIR)
	cp packages/multicaster/tokens.txt $(MULTICASTER_PKGDIR)

	$(call COPYTREE, packages/multicaster/service, $(MULTICASTER_PKGDIR)/service)

	mkdir -p $(MULTICASTER_PKGDIR)/bin/
	$(MAKE) -C packages/multicaster/src install DESTDIR=$(CURDIR)/$(MULTICASTER_PKGDIR)

multicaster-clean:
	rm -rf $(MULTICASTER_PKGDIR)
	$(MAKE) -C packages/multicaster/src clean

multicaster-build:
	$(MAKE) -C packages/multicaster/src build

PACKAGES += multicaster
