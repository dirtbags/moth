RADIO_PKGDIR = $(TARGET)/radio

radio-build:

radio-install:
	mkdir -p $(RADIO_PKGDIR)

	$(call COPYTREE, packages/radio/www, $(RADIO_PKGDIR)/www)

radio-clean:
	rm -rf $(RADIO_BUILDDIR)

PACKAGES += radio