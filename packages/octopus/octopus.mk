OCTOPUS_PKGDIR = $(TARGET)/octopus

octopus-install: octopus-build
	mkdir -p $(OCTOPUS_PKGDIR)/bin/

	$(call COPYTREE, packages/octopus/service, $(OCTOPUS_PKGDIR)/service)

	cp packages/octopus/tokens.txt $(OCTOPUS_PKGDIR)/
	cp packages/octopus/src/octopus $(OCTOPUS_PKGDIR)/bin/

octopus-clean:
	rm -rf $(OCTOPUS_PKGDIR)
	$(MAKE) -C packages/octopus/src clean

octopus-build:
	$(MAKE) -C packages/octopus/src build

PACKAGES += octopus
