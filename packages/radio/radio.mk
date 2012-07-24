RADIO_PKGDIR = $(TARGET)/radio

radio-install:
	mkdir -p $(RADIO_PKGDIR)
	cp packages/radio/answers.txt $(RADIO_PKGDIR)

PACKAGES += radio

