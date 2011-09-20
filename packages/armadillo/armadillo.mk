ARMADILLO_PKGDIR = $(TARGET)/armadillo

armadillo-install: armadillo-build
	mkdir -p $(ARMADILLO_PKGDIR)

	mkdir -p $(ARMADILLO_PKGDIR)/bin/
	$(MAKE) -C packages/armadillo/src install DESTDIR=$(CURDIR)/$(ARMADILLO_PKGDIR)

	$(call COPYTREE, packages/armadillo/tokens, $(ARMADILLO_PKGDIR)/tokens)	

	$(call COPYTREE, packages/armadillo/service, $(ARMADILLO_PKGDIR)/service)

armadillo-clean:
	rm -rf $(ARMADILLO_PKGDIR)
	$(MAKE) -C packages/armadillo/src clean

armadillo-build:
	$(MAKE) -C packages/armadillo/src build

PACKAGES += armadillo
