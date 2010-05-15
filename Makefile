BASE = /opt/ctf
VAR = $(BASE)/var
WWW = $(BASE)/www
LIB = $(BASE)/lib
BIN = $(BASE)/bin
SBIN = $(BASE)/sbin
BASE_URL = /

PYTHON = python
BUILD_DIR = build

TEMPLATE = $(CURDIR)/template.html
MDWNTOHTML = $(CURDIR)/mdwntohtml.py --template=$(TEMPLATE) --base=$(BASE_URL)

default: install

TARGETS = tanks puzzles
include $(wildcard */*.mk)
CLEAN_TARGETS = $(addsuffix -clean, $(TARGETS))
INSTALL_TARGETS = $(addsuffix -install, $(TARGETS))
.PHONY: $(CLEAN_TARGETS) $(INSTALL_TARGETS)

puzzles: puzzles/.git
puzzles/.git:
	git submodule update --init

puzzles-build: puzzles
	mkdir -p $(BUILD_DIR)/puzzles
	$(MAKE) -C puzzles BUILD_DIR=$(abspath $(BUILD_DIR)/puzzles)

puzzles-install: puzzles-build
	./mkpuzzles.py --base=$(BASE_URL) --puzzles=$(BUILD_DIR)/puzzles \
		--htmldir=$(DESTDIR)$(WWW)/puzzler \
		--keyfile=$(DESTDIR)$(LIB)/puzzler.keys

puzzles-clean:
	rm -rf $(BUILD_DIR)/puzzles $(DESTDIR)$(WWW)/puzzler $(DESTDIR)$(LIB)/puzzler.keys

tanks-install:
	install --directory $(DESTDIR)$(VAR)/tanks
	install --directory $(DESTDIR)$(VAR)/tanks/results
	install --directory $(DESTDIR)$(VAR)/tanks/errors
	install --directory $(DESTDIR)$(VAR)/tanks/ai
	install --directory $(DESTDIR)$(VAR)/tanks/ai/players
	install --directory $(DESTDIR)$(VAR)/tanks/ai/house

	ln -sf $(VAR)/tanks/results $(DESTDIR)$(WWW)/tanks/results

	install bin/run-tanks $(DESTDIR)$(SBIN)

tanks-clean:
	rm -rf $(DESTDIR)$(VAR)/tanks
	rm -rf $(DESTDIR)$(WWW)/tanks

install: $(INSTALL_TARGETS)
	install bin/pointscli $(DESTDIR)$(BIN)
	install bin/in.pointsd bin/in.flagd \
		bin/scoreboard \
		bin/run-ctf $(DESTDIR)$(SBIN)
	cp -r lib/* $(DESTDIR)$(LIB)
	cp -r www/* $(DESTDIR)$(WWW)
	cp template.html $(DESTDIR)$(LIB)

	install --directory $(DESTDIR)$(VAR)/disabled

	$(PYTHON) setup.py install --prefix=$(BASE)


$(INSTALL_TARGETS): base-install
base-install:
	install --directory $(DESTDIR)$(LIB) $(DESTDIR)$(BIN) $(DESTDIR)$(SBIN)
	install --directory $(DESTDIR)$(VAR)
	install --directory $(DESTDIR)$(WWW)
	install --directory $(DESTDIR)$(WWW)/puzzler
	install --directory $(DESTDIR)$(VAR)/points
	install --directory $(DESTDIR)$(VAR)/points/tmp
	install --directory $(DESTDIR)$(VAR)/points/cur
	install --directory $(DESTDIR)$(VAR)/flags

	echo 'VAR = "$(VAR)"' > ctf/paths.py
	echo 'WWW = "$(WWW)"' >> ctf/paths.py
	echo 'LIB = "$(LIB)"' >> ctf/paths.py
	echo 'BIN = "$(BIN)"' >> ctf/paths.py
	echo 'SBIN = "$(SBIN)"' >> ctf/paths.py
	echo 'BASE_URL = "$(BASE_URL)"' >> ctf/paths.py

uninstall:
	rm -rf $(DESTDIR)$(VAR) $(DESTDIR)$(WWW) $(DESTDIR)$(LIB)
	rm -rf $(DESTDIR)$(BIN) $(DESTDIR)$(SBIN)
	rmdir $(DESTDIR)$(BASE) || true


clean: $(CLEAN_TARGETS)
	$(MAKE) -C puzzles BUILD_DIR=$(abspath $(BUILD_DIR)/puzzles) clean

