PRINTF_PKGDIR = $(BUILD)/printf

printf-install: printf-build
	mkdir -p $(PRINTF_PKGDIR)

	mkdir -p $(PRINTF_PKGDIR)/bin/
	$(MAKE) -C packages/printf/src install DESTDIR=$(CURDIR)/$(PRINTF_PKGDIR)

	$(call COPYTREE, packages/printf/tokens, $(PRINTF_PKGDIR)/tokens)	

	$(call COPYTREE, packages/printf/service, $(PRINTF_PKGDIR)/service)

printf-clean:
	rm -rf $(PRINTF_PKGDIR)
	$(MAKE) -C packages/printf/src clean

printf-build:
	$(MAKE) -C packages/printf/src build

PACKAGES += printf
