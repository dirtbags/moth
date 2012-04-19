00ADMIN_PKGDIR = $(TARGET)/00admin

00admin-install:
	$(call COPYTREE, packages/00admin/service, $(00ADMIN_PKGDIR)/service)
	mkdir -p $(00ADMIN_PKGDIR)/sbin
	cp packages/00admin/sbin/* $(00ADMIN_PKGDIR)/sbin

PACKAGES += 00admin
