MCP_PKGDIR = $(TARGET)/mcp
MCP_BUILDDIR = $(BUILD)/mcp

MCP_FNORD_VERSION = 1.10
MCP_FNORD_TARBALL = fnord-$(MCP_FNORD_VERSION).tar.bz2
MCP_FNORD_TARCACHE = $(CACHE)/$(MCP_FNORD_TARBALL)
MCP_FNORD_URL = http://www.fefe.de/fnord/$(MCP_FNORD_TARBALL)
MCP_FNORD_SRCDIR = $(MCP_BUILDDIR)/fnord-$(MCP_FNORD_VERSION)

$(MCP_FNORD_TARCACHE):
	@ mkdir -p $(@D)
	wget -O $@ $(MCP_FNORD_URL)

mcp-source: $(MCP_BUILDDIR)/source
$(MCP_BUILDDIR)/source: $(MCP_FNORD_TARCACHE)
	mkdir -p $(@D)
	bzcat $< | (cd $(@D) && tar xf -)
	(cd $(@D)/fnord-$(MCP_FNORD_VERSION) && patch -p 1) < packages/mcp/fnord.patch
	touch $@

mcp-build: $(MCP_BUILDDIR)/build
$(MCP_BUILDDIR)/build: $(MCP_BUILDDIR)/source
	$(MAKE) -C packages/mcp/src build
	$(MAKE) -C $(MCP_BUILDDIR)/fnord-$(MCP_FNORD_VERSION) DIET= CC=$(CC) fnord-cgi


mcp-install: $(MCP_BUILDDIR)/build
	mkdir -p $(MCP_PKGDIR)

	$(call COPYTREE, packages/mcp/bin, $(MCP_PKGDIR)/bin)
	cp packages/mcp/src/pointscli $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/

	cp $(MCP_BUILDDIR)/fnord-$(MCP_FNORD_VERSION)/fnord-cgi $(MCP_PKGDIR)/bin/

	$(call COPYTREE, packages/mcp/service, $(MCP_PKGDIR)/service)

	$(call COPYTREE, packages/mcp/www, $(MCP_PKGDIR)/www)
	cp packages/mcp/src/puzzler.cgi $(MCP_PKGDIR)/www/
	cp packages/mcp/src/claim.cgi $(MCP_PKGDIR)/www/

mcp-test: mcp-build
	packages/mcp/test.sh

mcp-clean:
	rm -rf $(MCP_PKGDIR) $(MCP_BUILDDIR)
	$(MAKE) -C packages/mcp/src clean

PACKAGES += mcp