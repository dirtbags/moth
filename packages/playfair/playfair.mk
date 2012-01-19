PLAYFAIR_PKGDIR = $(TARGET)/playfair
PLAYFAIR_BUILDDIR = $(BUILD)/playfair

playfair-install: playfair-build
	mkdir -p $(PLAYFAIR_PKGDIR)/bin/

	$(call COPYTREE, packages/playfair/service, $(PLAYFAIR_PKGDIR)/service)

	cp packages/playfair/tokens.txt $(PLAYFAIR_PKGDIR)/
	cp packages/playfair/src/playfair $(PLAYFAIR_PKGDIR)/bin/

playfair-clean:
	rm -rf $(PLAYFAIR_PKGDIR) $(PLAYFAIR_BUILDDIR)
	$(MAKE) -C packages/playfair/src clean

playfair-build:
	$(MAKE) -C packages/playfair/src build

PACKAGES += playfair
