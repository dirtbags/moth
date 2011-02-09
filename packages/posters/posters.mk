posters-source:
posters-build:

posters-install: packages/posters/tokens.txt
	cp $< $(TARGET)/posters/
