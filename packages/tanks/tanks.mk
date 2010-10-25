TANKS_PKGDIR = $(TARGET)/tanks
TANKS_BUILDDIR = $(BUILD)/tanks
TANKS_TAR = $(CACHE)/tanks.tar.gz
TANKS_URL = "http://woozle.org/~neale/gitweb.cgi?p=ctanks;a=snapshot;h=master;sf=tgz"

$(TANKS_TAR):
	@ mkdir -p $(@D)
	wget -O $@ $(TANKS_URL)

tanks-source: $(TANKS_BUILDDIR)/ctanks
$(TANKS_BUILDDIR)/ctanks: $(TANKS_TAR)
	mkdir -p $(TANKS_BUILDDIR)
	zcat $(TANKS_TAR) | (cd $(TANKS_BUILDDIR) && tar xf -)

tanks-build: tanks-source
	$(MAKE) -C $(TANKS_BUILDDIR)/ctanks

tanks-install: tanks-build

tanks-clean:
	rm -f $(TANKS_BUILDDIR)
