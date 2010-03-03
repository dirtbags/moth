MDWN_DIR = mdwn

MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*.mdwn)
MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*/*.mdwn)
MDWN_SRC += $(wildcard $(MDWN_DIR)/src/*/*/*.mdwn)

MDWN_OUT = $(subst $(MDWN_DIR)/src/, $(WWW)/, $(MDWN_SRC:.mdwn=.html))

mdwn-install: $(MDWN_OUT)

$(WWW)/%.html: $(MDWN_DIR)/src/%.mdwn
	install -d $(@D)
	$(MDWNTOHTML) $< $@

mdwn-clean:
	rm -f $(MDWN_OUT)