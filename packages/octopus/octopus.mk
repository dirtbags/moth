OCTOPUS_PKGDIR = $(TARGET)/octopus

octopus-install: octopus-build
	mkdir -p $(OCTOPUS_PKGDIR)/bin/

	$(call COPYTREE, packages/octopus/service, $(OCTOPUS_PKGDIR)/service)

	$(call COPYTREE, packages/octopus/tokens, $(OCTOPUS_PKGDIR)/tokens)

	cp packages/octopus/src/octopus $(OCTOPUS_PKGDIR)/bin/

octopus-clean:
	rm -rf $(OCTOPUS_PKGDIR)
	$(MAKE) -C packages/octopus/src clean

octopus-build:
	$(MAKE) -C packages/octopus/src build

PACKAGES += octopus
