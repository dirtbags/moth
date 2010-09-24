MCP_PKGDIR = build/mcp
MCP_PACKAGE = mcp.pkg

mcp-package: $(MCP_PACKAGE)

$(MCP_PACKAGE): mcp-build
	mkdir -p $(MCP_PKGDIR)

	cp mcp/setup $(MCP_PKGDIR)

	$(call COPYTREE, mcp/bin, $(MCP_PKGDIR)/bin)
	cp mcp/src/in.tokend $(MCP_PKGDIR)/bin/
	cp mcp/src/tokencli $(MCP_PKGDIR)/bin/
	cp mcp/src/tokencli $(MCP_PKGDIR)/bin/
	cp mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/

	$(call COPYTREE, mcp/service, $(MCP_PKGDIR)/service)

	$(call COPYTREE, mcp/www, $(MCP_PKGDIR)/www)
	cp mcp/src/puzzler.cgi $(MCP_PKGDIR)/www/
	cp mcp/src/claim.cgi $(MCP_PKGDIR)/www/

	mksquashfs $(MCP_PKGDIR) $(MCP_PACKAGE) -all-root -noappend


mcp-test: mcp-build
	mcp/test.sh

mcp-clean:
	rm -rf $(MCP_PKGDIR) $(MCP_PACKAGE)
	$(MAKE) -C mcp/src clean

mcp-build:
	$(MAKE) -C mcp/src build

PACKAGES += mcp