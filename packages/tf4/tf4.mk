tf4-source:
tf4-build:

tf4-install: packages/tf4/tokens.txt
	mkdir -p $(TARGET)/tf4/
	cp $< $(TARGET)/tf4/

PACKAGES += tf4
