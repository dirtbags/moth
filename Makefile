BASE = /opt/ctf
VAR = $(BASE)/var
WWW = $(BASE)/www
LIB = $(BASE)/lib
BIN = $(BASE)/bin
SBIN = $(BASE)/sbin
BASE_URL = /

BUILD_DIR = build

TEMPLATE = $(CURDIR)/template.html
MDWNTOHTML = $(CURDIR)/mdwntohtml.py --template=$(TEMPLATE) --base=$(BASE_URL)

default: install

TARGETS = tanks puzzles
include $(wildcard */*.mk)
CLEAN_TARGETS = $(addsuffix -clean, $(TARGETS))
INSTALL_TARGETS = $(addsuffix -install, $(TARGETS))
.PHONY: $(CLEAN_TARGETS) $(INSTALL_TARGETS)

puzzles:
	git submodule update --init

puzzles-build: puzzles
	mkdir -p $(BUILD_DIR)/puzzles
	$(MAKE) -C puzzles BUILD_DIR=$(abspath $(BUILD_DIR)/puzzles)

puzzles-install: puzzles-build
	./mkpuzzles.py --base=$(BASE_URL) --puzzles=$(BUILD_DIR)/puzzles \
		--htmldir=$(WWW)/puzzler --keyfile=$(LIB)/puzzler.keys

puzzles-clean:
	rm -rf $(BUILD_DIR)/puzzles $(WWW)/puzzler $(LIB)/puzzler.keys

tanks-install:
	install --directory $(VAR)/tanks
	install --directory $(VAR)/tanks/results
	install --directory $(VAR)/tanks/errors
	install --directory $(VAR)/tanks/ai
	install --directory $(VAR)/tanks/ai/players
	install --directory $(VAR)/tanks/ai/house

	ln -sf $(VAR)/tanks/results $(WWW)/tanks/results

	install bin/run-tanks $(SBIN)

tanks-clean:
	rm -rf $(VAR)/tanks
	rm -rf $(WWW)/tanks

install: $(INSTALL_TARGETS)
	install bin/pointscli $(BIN)
	install bin/in.pointsd bin/in.flagd \
VAR		bin/scoreboard \
		bin/run-ctf $(SBIN)
	cp -r lib/* $(LIB)
	cp -r www/* $(WWW)
	cp template.html $(LIB)

	install --directory $(VAR)/disabled

	python setup.py install --prefix=$(BASE)


$(INSTALL_TARGETS): base-install
base-install:
	install --directory $(LIB) $(BIN) $(SBIN)
	install --directory $(VAR)
	install --directory $(WWW)
	install --directory $(WWW)/puzzler
	install --directory $(VAR)/points
	install --directory $(VAR)/points/tmp
	install --directory $(VAR)/points/cur
	install --directory $(VAR)/flags

	echo 'VAR = "$(VAR)"' > ctf/paths.py
	echo 'WWW = "$(WWW)"' >> ctf/paths.py
	echo 'LIB = "$(LIB)"' >> ctf/paths.py
	echo 'BIN = "$(BIN)"' >> ctf/paths.py
	echo 'SBIN = "$(SBIN)"' >> ctf/paths.py
	echo 'BASE_URL = "$(BASE_URL)"' >> ctf/paths.py

uninstall:
	rm -rf $(VAR) $(WWW) $(LIB) $(BIN) $(SBIN)
	rmdir $(BASE) || true


clean: $(CLEAN_TARGETS)
	$(MAKE) -C puzzles BUILD_DIR=$(abspath $(BUILD_DIR)/puzzles) clean

