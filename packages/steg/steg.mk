$(eval $(call STANDARD_PUZZLE, steg))

STEG_SUBMAKES = $(wildcard packages/steg/*/Makefile)
STEG_SUBCLEANS = $(patsubst %/Makefile, %/clean, $(STEG_SUBMAKES))

steg-clean: $(STEG_SUBCLEANS)

packages/steg/%/clean:
	$(MAKE) -C $(@D) clean
