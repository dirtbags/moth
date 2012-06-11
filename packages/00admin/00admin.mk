00ADMIN_PKGDIR = $(TARGET)/00admin
00ADMIN_BUILDDIR = $(BUILD)/00admin

ifndef PASSWORD
$(error PASSWORD not defined)
endif

00admin-build: $(00ADMIN_BUILDDIR)/build
$(00ADMIN_BUILDDIR)/build:
	$(MAKE) -C packages/00admin/src
	
00admin-install: $(00ADMIN_BUILDDIR)/build
	$(call COPYTREE, packages/00admin/service, $(00ADMIN_PKGDIR)/service)
	echo "$(PASSWORD)" > $(00ADMIN_PKGDIR)/password
	mkdir -p $(00ADMIN_PKGDIR)/sbin
	cp packages/00admin/bin/* $(00ADMIN_PKGDIR)/bin
	cp packages/00admin/src/tea $(00ADMIN_PKGDIR)/bin

PACKAGES += 00admin
