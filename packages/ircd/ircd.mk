IRCD_PKGDIR = $(TARGET)/ircd
IRCD_BUILDDIR = $(BUILD)/ircd
IRCD_VERSION = 16
IRCD_TAR = $(CACHE)/ngircd-$(IRCD_VERSION).tar.gz
IRCD_URL = ftp://ftp.berlios.de/pub/ngircd/ngircd-$(IRCD_VERSION).tar.gz
IRCD_SRCDIR = $(IRCD_BUILDDIR)/ngircd-$(IRCD_VERSION)

# Prevents automake from mangling cross-compiled binary names
IRCD_CONF_OPT := --program-transform-name=

ircd-install: ircd-build

$(IRCD_TAR):
	@ mkdir -p $(@D)
	wget -O $@ $(IRCD_URL)

ircd-source: $(IRCD_BUILDDIR)/source
$(IRCD_BUILDDIR)/source: $(IRCD_TAR)
	mkdir -p $(IRCD_BUILDDIR)
	zcat $(IRCD_TAR) | (cd $(IRCD_BUILDDIR) && tar xf -)
	touch $@

ircd-build: $(IRCD_BUILDDIR)/built
$(IRCD_BUILDDIR)/built: $(IRCD_BUILDDIR)/source
	cd $(IRCD_SRCDIR) && ./configure $(IRCD_CONF_OPT)
	$(MAKE) -C $(IRCD_SRCDIR)
	touch $@

ircd-install: ircd-build
	mkdir -p $(IRCD_PKGDIR)/bin
	cp $(IRCD_SRCDIR)/src/ngircd/ngircd $(IRCD_PKGDIR)/bin

	$(call COPYTREE, packages/ngircd/service, $(IRCD_PKGDIR)/service)

ircd-clean:
	rm -rf $(IRCD_BUILDDIR)

PACKAGES += ircd