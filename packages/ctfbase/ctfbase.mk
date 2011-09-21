CTFBASE_PKGDIR = $(TARGET)/ctfbase

ctfbase-install: ctfbase-build
	mkdir -p $(CTFBASE_PKGDIR)/bin/
	cp packages/ctfbase/src/arc4 $(CTFBASE_PKGDIR)/bin/

	$(call COPYTREE, packages/ctfbase/service, $(CTFBASE_PKGDIR)/service)

ctfbase-clean:
	rm -rf $(CTFBASE_PKGDIR)
	$(MAKE) -C packages/ctfbase/src clean

ctfbase-build:
	$(MAKE) -C packages/ctfbase/src build

PACKAGES += ctfbase
