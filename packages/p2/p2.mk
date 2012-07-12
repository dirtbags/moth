P2_PKGDIR = $(TARGET)/p2

p2-build: packages/p2/src/modem
packages/p2/src/modem:
	$(MAKE) -C packages/p2/src

p2-install: packages/p2/src/modem eris ctfbase
	mkdir -p $(P2_PKGDIR)

	$(call CTFBASE_INSTALL, $(P2_PKGDIR))

	$(call COPYTREE, packages/p2/bin, $(P2_PKGDIR)/bin)

	cp $(ERIS_BIN) $(P2_PKGDIR)/bin/
	cp packages/p2/src/modem $(P2_PKGDIR)/bin/

	$(call COPYTREE, packages/p2/service, $(P2_PKGDIR)/service)

	$(call COPYTREE, packages/p2/www, $(P2_PKGDIR)/www)

p2-clean:
	$(MAKE) -C packages/p2/src clean

PACKAGES += p2

