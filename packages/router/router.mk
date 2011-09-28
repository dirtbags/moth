ROUTER_PKGDIR = $(TARGET)/router
ROUTER_BUILDDIR = $(BUILD)/router

router-source:

router-build:

router-install: router-build
	$(call COPYTREE, packages/router/service, $(ROUTER_PKGDIR)/service)

router-clean:
	rm -rf $(ROUTER_PKGDIR) $(ROUTER_BUILDDIR)


##
## radvd
##
RADVD_VERSION = 1.8.1
RADVD_TARBALL = $(CACHE)/radvd-$(RADVD_VERSION).tar.gz
RADVD_URL = http://www.litech.org/radvd/dist/radvd-$(RADVD_VERSION).tar.gz
RADVD_SRCDIR = $(ROUTER_BUILDDIR)/radvd-$(RADVD_VERSION)

$(RADVD_TARBALL):
	@ mkdir -p $(@D)
	wget -O $@ $(RADVD_URL)

router-source: $(ROUTER_BUILDDIR)/radvd-source
$(ROUTER_BUILDDIR)/radvd-source: $(RADVD_TARBALL)
	mkdir -p $(ROUTER_BUILDDIR)
	zcat $(RADVD_TARBALL) | (cd $(ROUTER_BUILDDIR) && tar xf -)
	touch $@

router-build: $(ROUTER_BUILDDIR)/radvd-build
$(ROUTER_BUILDDIR)/radvd-build: $(ROUTER_BUILDDIR)/radvd-source
	cd $(RADVD_SRCDIR) && ./configure $(CONFIG_XCOMPILE_FLAGS)
	$(MAKE) -C $(RADVD_SRCDIR)
	touch $@

router-install: radvd-install
radvd-install:
	mkdir -p $(ROUTER_PKGDIR)/bin
	cp $(RADVD_SRCDIR)/radvd $(ROUTER_PKGDIR)/bin/
	cp $(RADVD_SRCDIR)/radvdump $(ROUTER_PKGDIR)/bin/


##
## mrd6
##
MRD6_CACHE = $(CACHE)/mrd6.git
MRD6_BUILDDIR = $(ROUTER_BUILDDIR)/mrd6
MRD6_URL = https://github.com/hugosantos/mrd6.git

$(MRD6_CACHE):
	git clone --bare $(MRD6_URL) $@

router-source: $(MRD6_BUILDDIR)
$(MRD6_BUILDDIR): $(MRD6_CACHE)
	git clone $< $@

router-build: $(ROUTER_BUILDDIR)/mrd6-build
$(ROUTER_BUILDDIR)/mrd6-build: $(MRD6_BUILDDIR)
	$(MAKE) -C $(MRD6_BUILDDIR)
	touch $@

router-install: mrd6-install
mrd6-install:
	mkdir -p $(ROUTER_PKGDIR)/bin
	cp $(MRD6_BUILDDIR)/src/mrd $(ROUTER_PKGDIR)/bin


PACKAGES += router
