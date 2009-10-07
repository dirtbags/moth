CGI = cat.cgi

TARGET = $(CURDIR)/target

FAKE = fakeroot -s $(CURDIR)/fake -i $(CURDIR)/fake
INSTALL = $(FAKE) install

all: 99-pwnables.tce

99-pwnables.tce: target
	$(FAKE) sh -c 'cd target && tar -czf - --exclude=placeholder --exclude=*~ .' > $@

target:
	$(INSTALL) -d $(TARGET)

	(cd skel; tar cf - .) | (cd $(TARGET); tar xf -)

	$(MAKE) -C daemons TARGET=$(TARGET) install

	$(INSTALL) -d $(TARGET)/usr/lib/www
	$(INSTALL) $(CGI) $(TARGET)/usr/lib/www

	$(INSTALL) -D flag $(TARGET)/var/lib/tftp/flag
	$(INSTALL) -D flag $(TARGET)/var/lib/notes/flag
	$(INSTALL) -D flag $(TARGET)/home/flag/.plan


clean:
	rm -rf target
