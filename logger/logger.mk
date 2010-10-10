LOGGER_PKGDIR = build/logger
LOGGER_PACKAGE = logger.pkg

logger-install: logger-build
	mkdir -p $(LOGGER_PKGDIR)

	mkdir -p $(LOGGER_PKGDIR)/bin/
	$(MAKE) -C logger/src install DESTDIR=$(CURDIR)/$(LOGGER_PKGDIR)

	$(call COPYTREE, logger/tokens, $(LOGGER_PKGDIR)/tokens)	

	$(call COPYTREE, logger/service, $(LOGGER_PKGDIR)/service)

logger-clean:
	rm -rf $(LOGGER_PKGDIR) $(LOGGER_PACKAGE)
	$(MAKE) -C logger/src clean

logger-build:
	$(MAKE) -C logger/src build

PACKAGES += logger
