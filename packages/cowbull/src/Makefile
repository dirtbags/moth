DESTDIR ?= /tmp

build: cowd cowcli

cowcli: cowcli.c
	cc -o $@ -static $<
	strip $@
	upx $@

install: build
	cp cowd cowcli $(DESTDIR)/bin/

clean:
	rm -f cowd cowcli
