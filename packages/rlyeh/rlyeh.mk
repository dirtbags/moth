RLYEH_PKGDIR = $(TARGET)/rlyeh
RLYEH_BUILDDIR = $(BUILD)/rlyeh
RLYEH_CACHE = $(CACHE)/rlyeh.git
RLYEH_URL = "http://woozle.org/~neale/projects/rlyeh"

$(RLYEH_CACHE):
	git clone --bare $(RLYEH_URL) $@

rlyeh-source: $(RLYEH_BUILDDIR)
$(RLYEH_BUILDDIR): $(RLYEH_CACHE)
	git clone $< $@

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
