SBIN = in.fingerd in.noted
BIN = notecli

TARGET ?= $(CURDIR)/target
FAKE = fakeroot -s $(CURDIR)/fake -i $(CURDIR)/fake
INSTALL = $(FAKE) install

all: $(SBIN) $(BIN)

install: all
	$(INSTALL) -d $(TARGET)/usr/sbin
	$(INSTALL) -s $(SBIN) $(TARGET)/usr/sbin

	$(INSTALL) -d $(TARGET)/usr/bin
	$(INSTALL) -s $(BIN) $(TARGET)/usr/bin

	$(INSTALL) -d $(TARGET)/var/service/fingerd
	$(INSTALL) run.fingerd $(TARGET)/var/service/fingerd/run

	$(INSTALL) -d $(TARGET)/var/service/noted
	$(INSTALL) run.noted $(TARGET)/var/service/noted/run
