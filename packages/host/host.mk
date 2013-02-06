HOST_PKGDIR = $(TARGET)/host

host-install:
	mkdir -p $(HOST_PKGDIR)
	cp packages/host/answers.txt $(HOST_PKGDIR)

PACKAGES += host

