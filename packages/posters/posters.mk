posters-source:
posters-build:

posters-install: packages/posters/tokens.txt
	mkdir -p $(TARGET)/posters/
	cp $< $(TARGET)/posters/

PACKAGES += posters
