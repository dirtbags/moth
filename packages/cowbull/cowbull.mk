COWBULL_PKGDIR = $(TARGET)/cowbull

cowbull-install: cowbull-build
	mkdir -p $(COWBULL_PKGDIR)

	mkdir -p $(COWBULL_PKGDIR)/bin/
	cp packages/cowbull/src/cowd $(COWBULL_PKGDIR)/bin

	mkdir -p $(COWBULL_PKGDIR)/www/cowbull/
	cp packages/cowbull/www/moo.html $(COWBULL_PKGDIR)/www/cowbull/index.html
	cp packages/cowbull/src/cowcli $(COWBULL_PKGDIR)/www/cowbull/
	
	$(call COPYTREE, packages/cowbull/service, $(COWBULL_PKGDIR)/service)
	cp packages/cowbull/tokens.txt $(COWBULL_PKGDIR)/

cowbull-clean:
	rm -rf $(COWBULL_PKGDIR)
	$(MAKE) -C packages/cowbull/src clean

cowbull-build:
	$(MAKE) -C packages/cowbull/src build

PACKAGES += cowbull
