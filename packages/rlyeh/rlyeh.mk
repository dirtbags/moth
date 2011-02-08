RLYEH_PKGDIR = $(TARGET)/rlyeh
RLYEH_BUILDDIR = $(BUILD)/rlyeh
RLYEH_TAR = $(CACHE)/rlyeh.tar.gz
RLYEH_URL = "http://woozle.org/~neale/gitweb.cgi?p=rlyeh;a=snapshot;h=master;sf=tgz"

$(RLYEH_TAR):
	@ mkdir -p $(@D)
	wget -O $@ $(RLYEH_URL)

rlyeh-source: $(RLYEH_BUILDDIR)/rlyeh
$(RLYEH_BUILDDIR)/rlyeh: $(RLYEH_TAR)
	mkdir -p $(RLYEH_BUILDDIR)
	zcat $(RLYEH_TAR) | (cd $(RLYEH_BUILDDIR) && tar xf -)

rlyeh-build: rlyeh-source
	$(MAKE) -C $(RLYEH_BUILDDIR)/rlyeh

rlyeh-install: rlyeh-build
	mkdir -p $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/rlyeh/rlyeh $(RLYEH_PKGDIR)/bin

	$(call COPYTREE, packages/rlyeh/service, $(RLYEH_PKGDIR)/service)

rlyeh-clean:
	rm -rf $(RLYEH_BUILDDIR)

PACKAGES += rlyeh
