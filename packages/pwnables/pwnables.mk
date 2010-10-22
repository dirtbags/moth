PWNABLES_PKGDIR = $(BUILD)/pwnables

pwnables-install: pwnables-build
	mkdir -p $(PWNABLES_PKGDIR)

	mkdir -p $(PWNABLES_PKGDIR)/bin/
	$(MAKE) -C packages/pwnables/src install DESTDIR=$(CURDIR)/$(PWNABLES_PKGDIR)

	$(call COPYTREE, packages/pwnables/tokens, $(PWNABLES_PKGDIR)/tokens)	

	$(call COPYTREE, packages/pwnables/service, $(PWNABLES_PKGDIR)/service)

pwnables-clean:
	rm -rf $(PWNABLES_PKGDIR)
	$(MAKE) -C packages/pwnables/src clean

pwnables-build:
	$(MAKE) -C packages/pwnables/src build

PACKAGES += pwnables
