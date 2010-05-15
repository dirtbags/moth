MDWN_DIR = mdwn

MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*.mdwn)
MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*/*.mdwn)
MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*/*/*.mdwn)

MDWN_OUT = $(subst $(MDWN_DIR)/src/, $(DESTDIR)$(WWW)/, $(MDWN_SRC:.mdwn=.html))

mdwn:

mdwn-install: $(MDWN_OUT)

$(DESTDIR)$(WWW)/%.html: $(MDWN_DIR)/src/%.mdwn
	install -d $(@D)
	$(MDWNTOHTML) $< $@

mdwn-clean:
	rm -f $(MDWN_OUT)

TARGETS += mdwn
