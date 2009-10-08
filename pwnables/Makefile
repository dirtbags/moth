CGI = cat.cgi

TARGET = $(CURDIR)/target

FAKE = fakeroot -s $(CURDIR)/fake -i $(CURDIR)/fake
INSTALL = $(FAKE) install

all: pwnables.tce

pwnables.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@

target:
	$(INSTALL) -d $(TARGET)

	(cd skel; tar cf - .) | (cd $(TARGET); tar xf -)

	$(MAKE) -C daemons TARGET=$(TARGET) install

clean:
	rm -rf target pwnables.tce
