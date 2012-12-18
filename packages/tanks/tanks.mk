TANKS_PKGDIR = $(TARGET)/tanks
TANKS_CACHE = $(CACHE)/tanks.git
TANKS_BUILDDIR = $(BUILD)/tanks
TANKS_URL = "http://woozle.org/~neale/g.cgi/tanks"

$(TANKS_CACHE):
	git clone --bare $(TANKS_URL) $@

tanks-source: $(TANKS_BUILDDIR)
$(TANKS_BUILDDIR): $(TANKS_CACHE)
	git clone $< $@

tanks-build: tanks-source
	$(MAKE) -C $(TANKS_BUILDDIR)

# "html" instead of "www" to prevent automatic links
tanks-install: tanks-build
	mkdir -p $(TANKS_PKGDIR)/bin
	cp $(TANKS_BUILDDIR)/forftanks $(TANKS_PKGDIR)/bin
	cp $(TANKS_BUILDDIR)/designer.cgi $(TANKS_PKGDIR)/bin
	cp $(TANKS_BUILDDIR)/rank.awk $(TANKS_PKGDIR)/bin
	cp $(TANKS_BUILDDIR)/winner.awk $(TANKS_PKGDIR)/bin

	$(call COPYTREE, packages/tanks/html, $(TANKS_PKGDIR)/html)
	cp packages/mcp/www/ctf.css $(TANKS_PKGDIR)/html/style.css
	cp $(TANKS_BUILDDIR)/nav.html.inc $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/tanks.js $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/forf.html $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/intro.html $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/figures.js $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/procs.html $(TANKS_PKGDIR)/html
	cp $(TANKS_BUILDDIR)/designer.js $(TANKS_PKGDIR)/html

	$(call COPYTREE, packages/tanks/service, $(TANKS_PKGDIR)/service)

	$(call COPYTREE, $(TANKS_BUILDDIR)/examples, $(TANKS_PKGDIR)/examples)

tanks-clean:
	rm -rf $(TANKS_BUILDDIR)

PACKAGES += tanks
