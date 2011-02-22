fireeye-source:
fireeye-build:

fireeye-install: packages/fireeye/tokens.txt
	mkdir -p $(TARGET)/fireeye/
	cp $< $(TARGET)/fireeye/

PACKAGES += fireeye
