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

# "html" instead of "www" to prevent automatic links
rlyeh-install: rlyeh-build
	mkdir -p $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/crlyeh/forfrlyeh $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/crlyeh/designer.cgi $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/crlyeh/rank.awk $(RLYEH_PKGDIR)/bin
	cp $(RLYEH_BUILDDIR)/crlyeh/winners.awk $(RLYEH_PKGDIR)/bin

	$(call COPYTREE, packages/rlyeh/html, $(RLYEH_PKGDIR)/html)
	cp packages/mcp/www/ctf.css $(RLYEH_PKGDIR)/html/style.css
	cp packages/mcp/www/grunge.png $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/nav.html.inc $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/rlyeh.js $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/forf.html $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/intro.html $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/figures.js $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/procs.html $(RLYEH_PKGDIR)/html
	cp $(RLYEH_BUILDDIR)/crlyeh/designer.js $(RLYEH_PKGDIR)/html

	$(call COPYTREE, packages/rlyeh/service, $(RLYEH_PKGDIR)/service)

	$(call COPYTREE, $(RLYEH_BUILDDIR)/crlyeh/examples, $(RLYEH_PKGDIR)/examples)

rlyeh-clean:
	rm -rf $(RLYEH_BUILDDIR)

PACKAGES += rlyeh
