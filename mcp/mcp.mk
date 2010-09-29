MCP_PKGDIR = build/mcp
MCP_PACKAGE = mcp.pkg

mcp-install: mcp-build
	mkdir -p $(MCP_PKGDIR)

	$(call COPYTREE, mcp/bin, $(MCP_PKGDIR)/bin)
	cp mcp/src/in.tokend $(MCP_PKGDIR)/bin/
	cp mcp/src/tokencli $(MCP_PKGDIR)/bin/
	cp mcp/src/tokencli $(MCP_PKGDIR)/bin/
	cp mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/

	$(call COPYTREE, mcp/service, $(MCP_PKGDIR)/service)

	$(call COPYTREE, mcp/tokend.keys, $(MCP_PKGDIR)/tokend.keys)

	$(call COPYTREE, mcp/www, $(MCP_PKGDIR)/www)
	cp mcp/src/puzzler.cgi $(MCP_PKGDIR)/www/
	cp mcp/src/claim.cgi $(MCP_PKGDIR)/www/

	touch $@

mcp-test: mcp-build
	mcp/test.sh

mcp-clean:
	rm -rf $(MCP_PKGDIR) $(MCP_PACKAGE) mcp-install
	$(MAKE) -C mcp/src clean

mcp-build:
	$(MAKE) -C mcp/src build

PACKAGES += mcp