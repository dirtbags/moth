PACKAGES = 

define COPYTREE
	mkdir -p $(2)
	(cd $(1) && find . -not -name "*~" | cpio -o) | (cd $(2) && cpio -i)
endef

define STANDARD_PUZZLE
t=$(strip $1)
$t-install: $(TARGET)/$t
$(TARGET)/$t: packages/$t
	mkdir -p $(TARGET)/$t
	./mkpuzzles packages/$t $(TARGET)/$t

$t-clean: $t-stdclean
$t-stdclean:
	rm -rf $(TARGET)/$t $(BIN)/$t.pkg

PACKAGES += $t
endef

include packages/*/*.mk

# Things configure likes to see
CONFIG_XCOMPILE_FLAGS = --host=i386-linux --program-transform-name=

# Make foo depend on foo.pkg
$(foreach p, $(PACKAGES), $(eval $p: $(BIN)/$p.pkg))

packages: $(patsubst %, $(BIN)/%.pkg, $(PACKAGES))

packages-install: $(addsuffix -install, $(PACKAGES))

packages-clean: $(addsuffix -clean, $(PACKAGES))
	rm -rf $(TARGET) $(BIN)

$(foreach p, $(PACKAGES), $(eval $p-clean: $p-pkgclean))
%-pkgclean:
	rm -f $(BIN)/$*.pkg

$(BIN)/%.pkg: %-install
	@ mkdir -p $(@D)
	mksquashfs $(TARGET)/$* $@ -all-root -noappend -no-progress
