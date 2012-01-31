DESTDIR ?= /tmp

all: cowd cowcli

cowcli: CC=cc

install: all
	cp cowd cowcli $(DESTDIR)/bin/

clean:
	rm -f cowd cowcli
