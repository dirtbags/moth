PRINTF_PKGDIR = build/printf
PRINTF_PACKAGE = printf.pkg

printf-install: printf-build
	mkdir -p $(PRINTF_PKGDIR)

	mkdir -p $(PRINTF_PKGDIR)/bin/
	$(MAKE) -C printf/src install DESTDIR=$(CURDIR)/$(PRINTF_PKGDIR)

	$(call COPYTREE, printf/tokens, $(PRINTF_PKGDIR)/tokens)	

	$(call COPYTREE, printf/service, $(PRINTF_PKGDIR)/service)

printf-clean:
	rm -rf $(PRINTF_PKGDIR) $(PRINTF_PACKAGE)
	$(MAKE) -C printf/src clean

printf-build:
	$(MAKE) -C printf/src build

PACKAGES += printf
