DESTDIR ?= /tmp

build: cowd cowcli

cowcli: CC=cc

install: build
	cp cowd cowcli $(DESTDIR)/bin/

clean:
	rm -f cowd cowcli
