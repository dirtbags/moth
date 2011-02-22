solera-source:
solera-build:

solera-install: packages/solera/tokens.txt
	mkdir -p $(TARGET)/solera/
	cp $< $(TARGET)/solera/

PACKAGES += solera
