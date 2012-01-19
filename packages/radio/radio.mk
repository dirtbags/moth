RADIO_PKGDIR = $(TARGET)/radio

radio-build:

radio-install:
	mkdir -p $(RADIO_PKGDIR)

	$(call COPYTREE, packages/radio/www, $(RADIO_PKGDIR)/www)
	cp packages/radio/tokens.txt $(RADIO_PKGDIR)/

radio-clean:
	rm -rf $(RADIO_BUILDDIR)

PACKAGES += radio