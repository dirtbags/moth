OCTOPUS_PKGDIR = build/octopus
OCTOPUS_PACKAGE = octopus.pkg

octopus-install: octopus-build
	mkdir -p $(OCTOPUS_PKGDIR)/bin/

	$(call COPYTREE, octopus/service, $(OCTOPUS_PKGDIR)/service)

	$(call COPYTREE, octopus/tokens, $(OCTOPUS_PKGDIR)/tokens)

	cp octopus/src/octopus $(OCTOPUS_PKGDIR)/bin/

octopus-clean:
	rm -rf $(OCTOPUS_PKGDIR) $(OCTOPUS_PACKAGE)
	$(MAKE) -C octopus/src clean

octopus-build:
	$(MAKE) -C octopus/src build

PACKAGES += octopus
