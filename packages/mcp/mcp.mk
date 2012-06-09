MCP_PKGDIR = $(TARGET)/mcp
MCP_BUILDDIR = $(BUILD)/mcp

mcp-source: $(MCP_BUILDDIR)/source
$(MCP_BUILDDIR)/source:
	mkdir -p $(@D)
	touch $@

mcp-build: $(MCP_BUILDDIR)/build
$(MCP_BUILDDIR)/build: $(MCP_BUILDDIR)/source
	$(MAKE) -C packages/mcp/src build

mcp-install: $(MCP_BUILDDIR)/build eris
	mkdir -p $(MCP_PKGDIR)

	$(call COPYTREE, packages/mcp/bin, $(MCP_PKGDIR)/bin)
	cp packages/mcp/src/pointscli $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/tea $(MCP_PKGDIR)/bin/

	cp $(ERIS_BIN) $(MCP_PKGDIR)/bin/

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