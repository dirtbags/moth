FIZZBUZZ_PKGDIR = $(TARGET)/fizzbuzz
FIZZBUZZ_BUILDDIR = $(BUILD)/fizzbuzz

$(FIZZBUZZ_BUILDDIR)/token.enc: packages/fizzbuzz/tokens.txt
$(FIZZBUZZ_BUILDDIR)/token.enc: packages/fizzbuzz/fizzbuzz-client.sh
$(FIZZBUZZ_BUILDDIR)/token.enc: $(FIZZBUZZ_BUILDDIR)/fizzbuzz-native
	packages/fizzbuzz/fizzbuzz-client.sh | $(FIZZBUZZ_BUILDDIR)/fizzbuzz-native 3< packages/fizzbuzz/tokens.txt > $@

$(FIZZBUZZ_BUILDDIR)/fizzbuzz-native: packages/fizzbuzz/src/fizzbuzz.c
	@ mkdir -p $(@D)
	cc -o $@ $<

fizzbuzz-install: fizzbuzz-build
	mkdir -p $(FIZZBUZZ_PKGDIR)/bin/

	$(call COPYTREE, packages/fizzbuzz/service, $(FIZZBUZZ_PKGDIR)/service)

	cp packages/fizzbuzz/tokens.txt $(FIZZBUZZ_PKGDIR)/
	cp $(FIZZBUZZ_BUILDDIR)/token.enc $(FIZZBUZZ_PKGDIR)/
	cp packages/fizzbuzz/src/fizzbuzz $(FIZZBUZZ_PKGDIR)/bin/

fizzbuzz-clean:
	rm -rf $(FIZZBUZZ_PKGDIR) $(FIZZBUZZ_BUILDDIR)
	$(MAKE) -C packages/fizzbuzz/src clean

fizzbuzz-build: $(FIZZBUZZ_BUILDDIR)/token.enc
	$(MAKE) -C packages/fizzbuzz/src build

PACKAGES += fizzbuzz
