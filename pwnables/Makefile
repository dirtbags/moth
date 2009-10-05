CGI = cat.cgi

TARGET = $(CURDIR)/target

FAKE = fakeroot -s $(CURDIR)/fake -i $(CURDIR)/fake
INSTALL = $(FAKE) install

all: target

target:
	$(INSTALL) -d $(TARGET)
	$(MAKE) -C daemons TARGET=$(TARGET) install

	$(INSTALL) -D init/networking $(TARGET)/var/service/networking/run
	$(INSTALL) -D init/telnetd $(TARGET)/var/service/networking/run
	$(INSTALL) -D init/tftpd $(TARGET)/var/service/networking/run
	$(INSTALL) -d $(TARGET)/usr/lib/www

	$(INSTALL) $(CGI) $(TARGET)/usr/lib/www

	$(INSTALL) -D flag $(TARGET)/var/lib/tftp/flag
	$(INSTALL) -D flag $(TARGET)/var/lib/notes/flag
	$(INSTALL) -D flag $(TARGET)/var/lib//flag
	$(INSTALL) -D flag $(TARGET)/var/lib/tftp/flag


clean:
	rm -rf target
