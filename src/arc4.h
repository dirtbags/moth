#ifndef __ARC4_H__
#define __ARC4_H__

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>

#define ARC4_KEYLEN 256

struct arc4_ctx {
  uint8_t S[256];
  uint8_t i;
  uint8_t j;
};

/* Stream operations */
ssize_t
arc4_encrypt_stream(FILE *out, FILE *in,
                    const uint8_t *key, size_t keylen);
ssize_t
arc4_decrypt_stream(FILE *out, FILE *in,
                    const uint8_t *key, size_t keylen);


/* Auto-seeding Psuedo Random Number Generator */
void arc4_rand_seed(const uint8_t *seed, size_t seedlen);
uint8_t arc4_rand8();
uint32_t arc4_rand32();

/* Low-level operations */
void arc4_init(struct arc4_ctx *ctx, const uint8_t *key, size_t keylen);
uint8_t arc4_out(struct arc4_ctx *ctx);
void arc4_crypt(struct arc4_ctx *ctx,
                uint8_t *obuf, const uint8_t *ibuf, size_t buflen);
void arc4_crypt_buffer(const uint8_t *key, size_t keylen,
                       uint8_t *buf, size_t buflen);
void arc4_nonce(uint8_t *nonce, size_t noncelen, void *seed, size_t seedlen);


#endif
