PACKAGES = 

define COPYTREE
	mkdir -p $(2)
	(cd $(1) && find . -not -name "*~" | cpio -o) | (cd $(2) && cpio -i)
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
