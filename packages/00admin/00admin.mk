00ADMIN_PKGDIR = $(TARGET)/00admin

ifndef PASSWORD
$(error PASSWORD not defined)
endif

00admin-install:
	$(call COPYTREE, packages/00admin/service, $(00ADMIN_PKGDIR)/service)
	echo "$(PASSWORD)" > $(00ADMIN_PKGDIR)/password
	mkdir -p $(00ADMIN_PKGDIR)/sbin
	cp packages/00admin/sbin/* $(00ADMIN_PKGDIR)/sbin

PACKAGES += 00admin
