##
## This is a non-package, for building eris httpd, which
## several packages use.  Just depend on $(ERIS_BIN), and
## copy it wherever you want in your install rule.
##

ERIS_CACHE = $(CACHE)/eris.git
ERIS_BUILDDIR = $(BUILD)/eris
ERIS_URL = http://woozle.org/~neale/projects/eris

ERIS_BIN := $(ERIS_BUILDDIR)/eris

$(ERIS_CACHE):
	git clone --bare $(ERIS_URL) $@

$(ERIS_BUILDDIR): $(ERIS_CACHE)
	git clone $< $@

eris: $(ERIS_BIN)
$(ERIS_BIN): $(ERIS_BUILDDIR)
	make -C $<
	