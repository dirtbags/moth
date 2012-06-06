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
RADVD_VERSION = 1.8.4
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
## ecmh
##
ECMH_CACHE = $(CACHE)/ecmh.git
ECMH_BUILDDIR = $(ROUTER_BUILDDIR)/ecmh
ECMH_URL = http://woozle.org/~neale/projects/ecmh

$(ECMH_CACHE):
	git clone --bare $(ECMH_URL) $@

router-source: $(ECMH_BUILDDIR)
$(ECMH_BUILDDIR): $(ECMH_CACHE)
	git clone $< $@

router-build: $(ROUTER_BUILDDIR)/ecmh-build
$(ROUTER_BUILDDIR)/ecmh-build: $(ECMH_BUILDDIR)
	$(MAKE) -C $(ECMH_BUILDDIR)/src ECMH_VERSION=dbtl-git STRIP=echo
	$(MAKE) -C $(ECMH_BUILDDIR)/tools/mtrace6
	touch $@

router-install: ecmh-install
ecmh-install:
	mkdir -p $(ROUTER_PKGDIR)/bin
	cp $(ECMH_BUILDDIR)/src/ecmh $(ROUTER_PKGDIR)/bin
	cp $(ECMH_BUILDDIR)/tools/mtrace6/mtrace6 $(ROUTER_PKGDIR)/bin

PACKAGES += router
