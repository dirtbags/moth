DESTDIR ?= /tmp

build: cowd cowcli

cowcli: cowcli.c
	$(CC) -o $@ -static $<
	strip $@
	upx $@

install: build
	cp cowd cowcli $(DESTDIR)/bin/

clean:
	rm -f cowd cowcli
