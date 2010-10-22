MCP_PKGDIR = $(BUILD)/mcp

mcp-install: mcp-build
	mkdir -p $(MCP_PKGDIR)

	$(call COPYTREE, packages/mcp/bin, $(MCP_PKGDIR)/bin)
	cp packages/mcp/src/in.tokend $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/pointscli $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/tokencli $(MCP_PKGDIR)/bin/
	cp packages/mcp/src/puzzles.cgi $(MCP_PKGDIR)/bin/

	$(call COPYTREE, packages/mcp/service, $(MCP_PKGDIR)/service)

	$(call COPYTREE, packages/mcp/tokend.keys, $(MCP_PKGDIR)/tokend.keys)

	$(call COPYTREE, packages/mcp/www, $(MCP_PKGDIR)/www)
	cp packages/mcp/src/puzzler.cgi $(MCP_PKGDIR)/www/
	cp packages/mcp/src/claim.cgi $(MCP_PKGDIR)/www/

mcp-test: mcp-build
	packages/mcp/test.sh

mcp-clean:
	rm -rf $(MCP_PKGDIR)
	$(MAKE) -C packages/mcp/src clean

mcp-build:
	$(MAKE) -C packages/mcp/src build

PACKAGES += mcp