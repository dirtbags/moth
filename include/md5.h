#ifndef MD5_H
#define MD5_H

#include <stdint.h>

/*  The following tests optimise behaviour on little-endian
    machines, where there is no need to reverse the byte order
    of 32 bit words in the MD5 computation.  By default,
    HIGHFIRST is defined, which indicates we're running on a
    big-endian (most significant byte first) machine, on which
    the byteReverse function in md5.c must be invoked. However,
    byteReverse is coded in such a way that it is an identity
    function when run on a little-endian machine, so calling it
    on such a platform causes no harm apart from wasting time.
    If the platform is known to be little-endian, we speed
    things up by undefining HIGHFIRST, which defines
    byteReverse as a null macro.  Doing things in this manner
    insures we work on new platforms regardless of their byte
    order.  */

#define HIGHFIRST

#ifdef __i386__
#undef HIGHFIRST
#endif

#define MD5_DIGEST_LEN 16
#define MD5_HEXDIGEST_LEN (MD5_DIGEST_LEN * 2)

struct md5_context {
  uint32_t buf[4];
  uint32_t bits[2];
  uint8_t in[64];
};

void md5_init(struct md5_context *ctx);
void md5_update(struct md5_context *ctx, const uint8_t *buf, size_t len);
void md5_final(struct md5_context *ctx, uint8_t *digest);
void md5_digest(const uint8_t *buf, size_t buflen, uint8_t *digest);
void md5_hexdigest(const uint8_t *buf, size_t buflen, char *hexdigest);

#endif /* !MD5_H */
