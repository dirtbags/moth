TOKENCLI_PKGDIR = build/tokencli
TOKENCLI_PACKAGE = tokencli.pkg

tokencli-install: tokencli-build
	mkdir -p $(TOKENCLI_PKGDIR)/bin/

	$(call COPYTREE, tokencli/service, $(TOKENCLI_PKGDIR)/service)

	cp tokencli/setup $(TOKENCLI_PKGDIR)/

	cp tokencli/src/tokencli $(TOKENCLI_PKGDIR)/bin/
	cp tokencli/src/arc4 $(TOKENCLI_PKGDIR)/bin/

tokencli-clean:
	rm -rf $(TOKENCLI_PKGDIR) $(TOKENCLI_PACKAGE)
	$(MAKE) -C tokencli/src clean

tokencli-build:
	$(MAKE) -C tokencli/src build

PACKAGES += tokencli
