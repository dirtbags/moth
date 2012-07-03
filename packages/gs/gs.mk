GS_PKGDIR = $(TARGET)/gs

gs-install:
	mkdir -p $(GS_PKGDIR)
	cp packages/gs/answers.txt $(GS_PKGDIR)

PACKAGES += gs

