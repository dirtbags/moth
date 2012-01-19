LXC_PKGDIR = $(TARGET)/lxc
LXC_BUILDDIR = $(BUILD)/lxc
LXC_VERSION = 0.7.5
LXC_TAR = $(CACHE)/lxc-$(LXC_VERSION).tar.gz
LXC_URL = http://lxc.sourceforge.net/download/lxc/lxc-$(LXC_VERSION).tar.gz
LXC_SRCDIR = $(LXC_BUILDDIR)/lxc-$(LXC_VERSION)

LXC_COMMANDS  = attach cgroup checkpoint console execute freeze
LXC_COMMANDS += info init kill monitor restart start stop
LXC_COMMANDS += unfreeze unshare wait

LXC_PROGRAMS = $(addprefix $(LXC_SRCDIR)/src/lxc/lxc-, $(LXC_COMMANDS))


# Prevents automake from mangling cross-compiled binary names
LXC_CC_HOST := $(shell $(CC) -v 2>&1 | awk '/Target:/{print $$2}')
LXC_CONF_OPT := --host=i386-unknown-linux-uclibc --program-transform-name=

lxc-install: lxc-build

$(LXC_TAR):
	@ mkdir -p $(@D)
	wget -O $@ $(LXC_URL)

lxc-source: $(LXC_BUILDDIR)/source
$(LXC_BUILDDIR)/source: $(LXC_TAR)
	mkdir -p $(LXC_BUILDDIR)
	zcat $(LXC_TAR) | (cd $(LXC_BUILDDIR) && tar xf -)
	cp packages/lxc/utmp.c $(LXC_SRCDIR)/src/lxc/
	touch $@

lxc-build: $(LXC_BUILDDIR)/built
$(LXC_BUILDDIR)/built: $(LXC_BUILDDIR)/source libcap-build
	cd $(LXC_SRCDIR) && CFLAGS="$(LIBCAP_CFLAGS)" LDFLAGS="$(LIBCAP_LDFLAGS) -Xlinker -rpath -Xlinker /opt/lxc/lib" ./configure $(CONFIG_XCOMPILE_FLAGS)
	$(MAKE) -C $(LXC_SRCDIR)
	touch $@

lxc-install: lxc-build
	mkdir -p $(LXC_PKGDIR)/lib
	cp $(LXC_SRCDIR)/src/lxc/liblxc.so $(LXC_PKGDIR)/lib/liblxc.so.0
	cp $(LIBCAP_SRCDIR)/libcap/libcap.so.* $(LXC_PKGDIR)/lib

	mkdir -p $(LXC_PKGDIR)/bin
	cp $(LXC_PROGRAMS) $(LXC_PKGDIR)/bin

#	$(call COPYTREE, packages/lxc/service, $(LXC_PKGDIR)/service)

lxc-clean:
	rm -rf $(LXC_BUILDDIR)


LIBCAP_PKGDIR = $(TARGET)/libcap
