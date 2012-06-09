INFERNO_PKGDIR = $(TARGET)/inferno
INFERNO_BUILDDIR = $(BUILD)/inferno

inferno-source:

inferno-build: 

inferno-install: eris
	mkdir -p $(INFERNO_PKGDIR)/bin

	cp $(ERIS_BIN) $(INFERNO_PKGDIR)/bin/

	$(call COPYTREE, packages/inferno/service, $(INFERNO_PKGDIR)/service)

inferno-clean:
	rm -rf $(INFERNO_PKGDIR) $(INFERNO_BUILDDIR) 

PACKAGES += inferno
