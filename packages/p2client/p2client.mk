P2CLIENT_PKGDIR = $(TARGET)/p2client

p2client-source:
p2client-build:
p2client-install:
	mkdir -p $(P2CLIENT_PKGDIR)

	$(call COPYTREE, packages/p2client/service, $(P2CLIENT_PKGDIR)/service)
	loadkeys -b packages/p2client/dumbterm.map > $(P2CLIENT_PKGDIR)/dumbterm.kmap
	cp packages/p2client/lite-16.fnt $(P2CLIENT_PKGDIR)

p2client-clean:
	rm -rf $(P2CLIENT_PKGDIR)

PACKAGES += p2client