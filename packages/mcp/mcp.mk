MCP_PKGDIR = $(TARGET)/mcp
MCP_BUILDDIR = $(BUILD)/mcp


##
## XXX: clean up fnord like router.mk
##

FNORD_CACHE = $(CACHE)/fnord.git
FNORD_BUILDDIR = $(MCP_BUILDDIR)/fnord
FNORD_URL = http://woozle.org/~neale/projects/fnord

$(FNORD_CACHE):
	git clone --bare $(FNORD_URL) $@

mcp-source: $(MCP_BUILDDIR)/source
$(MCP_BUILDDIR)/source: $(FNORD_CACHE)
	mkdir -p $(@D)
	git clone $(FNORD_CACHE) $(@D)/fnord
	touch $@

mcp-build: $(MCP_BUILDDIR)/build
$(MCP_BUILDDIR)/build: $(MCP_BUILDDIR)/source
	$(MAKE) -C packages/mcp/src build
	$(MAKE) -C $(MCP_BUILDDIR)/fnord DIET= CC=$(CC) fnord-cgi


mcp-install: $(MCP_BUILDDIR)/build
	mkdir -p $(MCP_PKGDIR)

	$(call COPYTREE, packages/mcp/bin, $(MCP_PKGDIR)/bin)
	cp packages/mcp/src/pointscli $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/

	cp $(MCP_BUILDDIR)/fnord/fnord-cgi $(MCP_PKGDIR)/bin/

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