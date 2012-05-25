INFERNO_PKGDIR = $(TARGET)/inferno
INFERNO_BUILDDIR = $(BUILD)/inferno

INFERNO_ERIS_CACHE = $(CACHE)/eris.git
INFERNO_ERIS_URL = http://woozle.org/~neale/projects/eris

$(INFERNO_ERIS_CACHE):
	git clone --bare $(INFERNO_ERIS_URL) $@

inferno-source: $(INFERNO_BUILDDIR)
$(INFERNO_BUILDDIR): $(INFERNO_ERIS_CACHE)
	git clone $< $@

inferno-build: $(INFERNO_BUILDDIR)/build
$(INFERNO_BUILDDIR)/build: $(INFERNO_BUILDDIR)
	$(MAKE) -C $(INFERNO_BUILDDIR)

inferno-install: $(INFERNO_BUILDDIR)/build
	mkdir -p $(INFERNO_PKGDIR)/bin

	cp $(INFERNO_BUILDDIR)/eris $(INFERNO_PKGDIR)/bin/

	$(call COPYTREE, packages/inferno/service, $(INFERNO_PKGDIR)/service)

inferno-clean:
	rm -rf $(INFERNO_PKGDIR) $(INFERNO_BUILDDIR) 

PACKAGES += inferno
