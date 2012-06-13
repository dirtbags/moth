ifndef PASSWORD
$(error PASSWORD not defined)
endif

TEA_BIN = packages/00common/src/tea
POINTSCLI_BIN = packages/00common/src/pointscli
PUZZLES_BIN = packages/00common/src/puzzles.cgi
	
.PHONY: ctfbase
ctfbase: $(TEA_BIN) $(POINTSCLI_BIN) $(PUZZLES_BIN)
$(TEA_BIN) $(POINTSCLI_BIN) $(PUZZLES_BIN):
	$(MAKE) -C $(@D)

packages-clean: ctfbase-clean
ctfbase-clean:
	$(MAKE) -C packages/00common/src clean

define CTFBASE_INSTALL
	$(call COPYTREE, packages/00common/service, $1/service)

	mkdir -p $(1)/bin
	cp $(TEA_BIN) $(1)/bin
	cp $(POINTSCLI_BIN) $(1)/bin
	cp $(PUZZLES_BIN) $(1)/bin

	echo "$(PASSWORD)" > $(1)/password
endef
