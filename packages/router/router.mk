ROUTER_PKGDIR = $(TARGET)/router
ROUTER_BUILDDIR = $(BUILD)/router

DNSMASQ_VERSION = 2.57
DNSMASQ_SRCDIR = $(ROUTER_BUILDDIR)/dnsmasq-$(DNSMASQ_VERSION)
DNSMASQ_TARBALL = $(CACHE)/dnsmasq-$(DNSMASQ_VERSION).tar.gz
DNSMASQ_URL = http://www.thekelleys.org.uk/dnsmasq/dnsmasq-$(DNSMASQ_VERSION).tar.gz

$(DNSMASQ_TARBALL):
	@ mkdir -p $(@D)
	wget -O $@ $(DNSMASQ_URL)

router-source: $(ROUTER_BUILDDIR)/source
$(ROUTER_BUILDDIR)/source: $(DNSMASQ_TARBALL)
	mkdir -p $(ROUTER_BUILDDIR)
	zcat $(DNSMASQ_TARBALL) | (cd $(ROUTER_BUILDDIR) && tar xf -)
	touch $@

router-build: $(ROUTER_BUILDDIR)/built
$(ROUTER_BUILDDIR)/built: $(ROUTER_BUILDDIR)/source
	$(MAKE) -C $(DNSMASQ_SRCDIR)
	touch $@

router-install: router-build
	mkdir -p $(ROUTER_PKGDIR)/sbin
	cp $(DNSMASQ_SRCDIR)/src/dnsmasq $(ROUTER_PKGDIR)/sbin/

	$(call COPYTREE, packages/router/service, $(ROUTER_PKGDIR)/service)

router-clean:
	rm -rf $(ROUTER_PKGDIR)

PACKAGES += router
