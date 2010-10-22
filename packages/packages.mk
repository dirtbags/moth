PACKAGES = 

define COPYTREE
	mkdir -p $(2)
	(cd $(1) && find . -not -name "*~" | cpio -o) | (cd $(2) && cpio -i)
endef

define STANDARD_PUZZLE
t=$(strip $1)
$t-install:
	mkdir -p $(BUILD)/$t
	./mkpuzzles packages/$t $(BUILD)/$t

$t-clean:
	rm -rf $(BUILD)/$t $(BIN)/$t.pkg

PACKAGES += $t
endef

include packages/*/*.mk

# Make foo depend on foo.pkg
$(foreach p, $(PACKAGES), $(eval $p: $(BIN)/$p.pkg))

packages: $(patsubst %, $(BIN)/%.pkg, $(PACKAGES))

install: $(addsuffix -install, $(PACKAGES))

clean: $(addsuffix -clean, $(PACKAGES))
	rm -rf $(BUILD) $(BIN)

$(foreach p, $(PACKAGES), $(eval $p-clean: $p-pkgclean))
%-pkgclean:
	rm -f $(BIN)/$*.pkg

$(BIN)/%.pkg: %-install
	@ mkdir -p $(@D)
	mksquashfs $(BUILD)/$* $@ -all-root -noappend
