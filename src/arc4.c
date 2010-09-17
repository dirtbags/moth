#include <stdint.h>
#include <stdlib.h>
#include "arc4.h"

#define swap(a, b) do {int _swap=a; a=b, b=_swap;} while (0)

void
arc4_init(struct arc4_ctx *ctx, uint8_t const *key, size_t keylen)
{
  int i;
  int j = 0;

  for (i = 0; i < 256; i += 1) {
    ctx->S[i] = i;
  }

  for (i = 0; i < 256; i += 1) {
    j = (j + ctx->S[i] + key[i % keylen]) % 256;
    swap(ctx->S[i], ctx->S[j]);
  }
  ctx->i = 0;
  ctx->j = 0;
}

void
arc4_crypt(struct arc4_ctx *ctx,
           uint8_t *obuf, uint8_t const *ibuf, size_t buflen)
{
  int    i = ctx->i;
  int    j = ctx->j;
  size_t k;

  for (k = 0; k < buflen; k += 1) {
    uint8_t mask;

    i = (i + 1) % 256;
    j = (j + ctx->S[i]) % 256;
    swap(ctx->S[i], ctx->S[j]);
    mask = ctx->S[(ctx->S[i] + ctx->S[j]) % 256];
    obuf[k] = ibuf[k] ^ mask;
  }
  ctx->i = i;
  ctx->j = j;
}

void
arc4_crypt_buffer(uint8_t const *key, size_t keylen,
                  uint8_t *buf, size_t buflen)
{
  struct arc4_ctx ctx;

  arc4_init(&ctx, key, keylen);
  arc4_crypt(&ctx, buf, buf, buflen);
}
