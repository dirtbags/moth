LIBCAP_PKGDIR = $(TARGET)/libcap
LIBCAP_BUILDDIR = $(BUILD)/libcap
LIBCAP_VERSION = 2.22
LIBCAP_TAR = $(CACHE)/libcap-$(LIBCAP_VERSION).tar.gz
# XXX: kernel.org was down when I wrote this, but is the canonical source
LIBCAP_URL = http://ftp.debian.org/debian/pool/main/libc/libcap2/libcap2_$(LIBCAP_VERSION).orig.tar.gz
LIBCAP_SRCDIR = $(LIBCAP_BUILDDIR)/libcap-$(LIBCAP_VERSION)

LIBCAP_LDOPTS = -L$(CURDIR)/$(LIBCAP_SRCDIR)/libcap
LIBCAP_CFLAGS = -I$(CURDIR)/$(LIBCAP_SRCDIR)/libcap/include

$(LIBCAP_TAR):
	mkdir -p $(@D)
	wget -O $@ $(LIBCAP_URL)

libcap-source: $(LIBCAP_BUILDDIR)/source
$(LIBCAP_BUILDDIR)/source: $(LIBCAP_TAR)
	mkdir -p $(@D)
	zcat $< | ( cd $(@D) && tar xf -)
	touch $@


# This library's build sort of blows.
libcap-build: $(LIBCAP_BUILDDIR)/built
$(LIBCAP_BUILDDIR)/built: $(LIBCAP_BUILDDIR)/source
	$(MAKE) -C $(LIBCAP_SRCDIR)/libcap _makenames
	$(MAKE) -C $(LIBCAP_SRCDIR) CC=$(CC)
	touch $@
