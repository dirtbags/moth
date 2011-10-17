00ADMIN_PKGDIR = $(TARGET)/00admin

00admin-install:
	$(call COPYTREE, packages/00admin/service, $(00ADMIN_PKGDIR)/service)

PACKAGES += 00admin
