PWNABLES_PKGDIR = build/pwnables
PWNABLES_PACKAGE = pwnables.pkg

pwnables-install: pwnables-build
	mkdir -p $(PWNABLES_PKGDIR)

	cp pwnables/setup $(PWNABLES_PKGDIR)

	mkdir -p $(PWNABLES_PKGDIR)/bin/
	$(MAKE) -C pwnables/src install DESTDIR=$(CURDIR)/$(PWNABLES_PKGDIR)

	$(call COPYTREE, pwnables/service, $(PWNABLES_PKGDIR)/service)

pwnables-clean:
	rm -rf $(PWNABLES_PKGDIR) $(PWNABLES_PACKAGE)
	$(MAKE) -C pwnables/src clean

pwnables-build:
	$(MAKE) -C pwnables/src build

PACKAGES += pwnables
