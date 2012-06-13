MCP_PKGDIR = $(TARGET)/mcp

mcp-install: eris ctfbase
	mkdir -p $(MCP_PKGDIR)

	$(call CTFBASE_INSTALL, $(MCP_PKGDIR))

	$(call COPYTREE, packages/mcp/bin, $(MCP_PKGDIR)/bin)

	cp $(ERIS_BIN) $(MCP_PKGDIR)/bin/

	$(call COPYTREE, packages/mcp/service, $(MCP_PKGDIR)/service)

	$(call COPYTREE, packages/mcp/www, $(MCP_PKGDIR)/www)
	cp packages/00common/src/puzzler.cgi $(MCP_PKGDIR)/www/
	cp packages/00common/src/claim.cgi $(MCP_PKGDIR)/www/

mcp-test: mcp-build
	packages/mcp/test.sh

PACKAGES += mcp