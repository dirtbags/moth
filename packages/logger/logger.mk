LOGGER_PKGDIR = $(BUILD)/logger

logger-install: logger-build
	mkdir -p $(LOGGER_PKGDIR)

	mkdir -p $(LOGGER_PKGDIR)/bin/
	$(MAKE) -C packages/logger/src install DESTDIR=$(CURDIR)/$(LOGGER_PKGDIR)

	$(call COPYTREE, packages/logger/tokens, $(LOGGER_PKGDIR)/tokens)	

	$(call COPYTREE, packages/logger/service, $(LOGGER_PKGDIR)/service)

logger-clean:
	rm -rf $(LOGGER_PKGDIR)
	$(MAKE) -C packages/logger/src clean

logger-build:
	$(MAKE) -C packages/logger/src build

PACKAGES += logger
