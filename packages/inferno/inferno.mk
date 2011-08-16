INFERNO_PKGDIR = $(TARGET)/inferno
INFERNO_BUILDDIR = $(BUILD)/inferno

INFERNO_FNORD_CACHE = $(CACHE)/fnord.git
INFERNO_FNORD_URL = http://woozle.org/~neale/projects/fnord

$(INFERNO_FNORD_CACHE):
	git clone --bare $(INFERNO_FNORD_URL) $@

inferno-source: $(INFERNO_BUILDDIR)
$(INFERNO_BUILDDIR): $(INFERNO_FNORD_CACHE)
	git clone $< $@

inferno-build: $(INFERNO_BUILDDIR)/build
$(INFERNO_BUILDDIR)/build: $(INFERNO_BUILDDIR)
	$(MAKE) -C $(INFERNO_BUILDDIR) fnord-idx

inferno-install: $(INFERNO_BUILDDIR)/build
	mkdir -p $(INFERNO_PKGDIR)/bin

	cp $(INFERNO_BUILDDIR)/fnord-idx $(INFERNO_PKGDIR)/bin/

	$(call COPYTREE, packages/inferno/service, $(INFERNO_PKGDIR)/service)

inferno-clean:
	rm -rf $(INFERNO_PKGDIR) $(INFERNO_BUILDDIR) 
	$(MAKE) -C packages/inferno/src clean

PACKAGES += inferno