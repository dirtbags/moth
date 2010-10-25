TOKENS_PKGDIR = $(TARGET)/tokens

tokens-install: tokens-build
	mkdir -p $(TOKENS_PKGDIR)/bin/

	$(call COPYTREE, packages/tokens/service, $(TOKENS_PKGDIR)/service)

	cp packages/tokens/setup $(TOKENS_PKGDIR)/

	cp packages/tokens/src/tokencli $(TOKENS_PKGDIR)/bin/
	cp packages/tokens/src/arc4 $(TOKENS_PKGDIR)/bin/

tokens-clean:
	rm -rf $(TOKENS_PKGDIR)
	$(MAKE) -C packages/tokens/src clean

tokens-build:
	$(MAKE) -C packages/tokens/src build

PACKAGES += tokens
