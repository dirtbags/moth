$(eval $(call STANDARD_PUZZLE, sequence))

SEQUENCE_SUBMAKES = $(wildcard packages/sequence/*/Makefile)
SEQUENCE_SUBCLEANS = $(patsubst %/Makefile, %/clean, $(SEQUENCE_SUBMAKES))

sequence-clean: $(SEQUENCE_SUBCLEANS)

packages/sequence/%/clean:
	$(MAKE) -C $(@D) clean
