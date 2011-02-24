RLYEH_PKGDIR = $(TARGET)/rlyeh
RLYEH_BUILDDIR = $(BUILD)/rlyeh
RLYEH_URL = "http://woozle.org/~neale/projects/rlyeh"

rlyeh-source: $(RLYEH_BUILDDIR)
$(RLYEH_BUILDDIR):
	git clone $(RLYEH_URL) $@

rlyeh-build: rlyeh-source
	$(MAKE) -C $(RLYEH_BUILDDIR)

rlyeh-install: rlyeh-build
	mkdir -p $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/rlyeh $(RLYEH_PKGDIR)/bin

	$(call COPYTREE, packages/rlyeh/service, $(RLYEH_PKGDIR)/service)
	$(call COPYTREE, packages/rlyeh/tokens, $(RLYEH_PKGDIR)/tokens)

rlyeh-clean:
	rm -rf $(RLYEH_BUILDDIR)

PACKAGES += rlyeh
