sandia-source:
sandia-build:

sandia-install: packages/sandia/tokens.txt
	mkdir -p $(TARGET)/sandia/
	cp $< $(TARGET)/sandia/

PACKAGES += sandia
