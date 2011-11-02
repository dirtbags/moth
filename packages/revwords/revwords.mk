REVWORDS_PKGDIR = $(TARGET)/revwords
REVWORDS_BUILDDIR = $(BUILD)/revwords

revwords-install: revwords-build
	mkdir -p $(REVWORDS_PKGDIR)/bin/

	$(call COPYTREE, packages/revwords/service, $(REVWORDS_PKGDIR)/service)

	cp packages/revwords/tokens.txt $(REVWORDS_PKGDIR)/
	cp $(REVWORDS_BUILDDIR)/token.enc $(REVWORDS_PKGDIR)/
	cp packages/revwords/src/revwords $(REVWORDS_PKGDIR)/bin/

revwords-clean:
	rm -rf $(REVWORDS_PKGDIR) $(REVWORDS_BUILDDIR)
	$(MAKE) -C packages/revwords/src clean

revwords-build: $(REVWORDS_BUILDDIR)/token.enc
	$(MAKE) -C packages/revwords/src build

PACKAGES += revwords
