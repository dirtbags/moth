ROUTER_PKGDIR = $(TARGET)/router
ROUTER_BUILDDIR = $(BUILD)/router


router-source: $(ROUTER_BUILDDIR)/radvd-source

router-build: $(ROUTER_BUILDDIR)/radvd-build

router-install: router-build
	mkdir -p $(ROUTER_PKGDIR)/bin

	cp $(RADVD_SRCDIR)/radvd $(ROUTER_PKGDIR)/bin/
	cp $(RADVD_SRCDIR)/radvdump $(ROUTER_PKGDIR)/bin/

	$(call COPYTREE, packages/router/service, $(ROUTER_PKGDIR)/service)

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

$(ROUTER_BUILDDIR)/radvd-source: $(RADVD_TARBALL)
	mkdir -p $(ROUTER_BUILDDIR)
	zcat $(RADVD_TARBALL) | (cd $(ROUTER_BUILDDIR) && tar xf -)
	touch $@

$(ROUTER_BUILDDIR)/radvd-build: $(ROUTER_BUILDDIR)/radvd-source
	cd $(RADVD_SRCDIR) && ./configure $(CONFIG_XCOMPILE_FLAGS)
	$(MAKE) -C $(RADVD_SRCDIR)
	touch $@



router-clean:
	rm -rf $(ROUTER_PKGDIR)

PACKAGES += router
