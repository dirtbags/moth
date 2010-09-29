TOKENS_PKGDIR = build/tokens
TOKENS_PACKAGE = tokens.pkg

tokens-install: tokens-build
	mkdir -p $(TOKENS_PKGDIR)/bin/

	$(call COPYTREE, tokens/service, $(TOKENS_PKGDIR)/service)

	cp tokens/setup $(TOKENS_PKGDIR)/

	cp tokens/src/tokencli $(TOKENS_PKGDIR)/bin/
	cp tokens/src/arc4 $(TOKENS_PKGDIR)/bin/

tokens-clean:
	rm -rf $(TOKENS_PKGDIR) $(TOKENS_PACKAGE)
	$(MAKE) -C tokens/src clean

tokens-build:
	$(MAKE) -C tokens/src build

PACKAGES += tokens
