HOST_PKGDIR = $(TARGET)/host

host-install:
	mkdir -p $(HOST_PKGDIR)
	cp packages/host/tokens.txt $(HOST_PKGDIR)

PACKAGES += host

