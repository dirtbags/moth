WOPR_PKGDIR = $(TARGET)/wopr

wopr-source:
wopr-build:

wopr-install: packages/wopr/tokens.txt
	mkdir -p $(WOPR_PKGDIR)
	cp packages/wopr/tokens.txt $(WOPR_PKGDIR)/

	$(call COPYTREE, packages/wopr/www, $(WOPR_PKGDIR)/www)
	$(call COPYTREE, packages/wopr/service, $(WOPR_PKGDIR)/service)

wopr-clean:

PACKAGES += wopr
