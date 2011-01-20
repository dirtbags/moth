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

uint8_t
arc4_pad(struct arc4_ctx *ctx)
{
  ctx->i = (ctx->i + 1) % 256;
  ctx->j = (ctx->j + ctx->S[ctx->i]) % 256;
  swap(ctx->S[ctx->i], ctx->S[ctx->j]);
  return ctx->S[(ctx->S[ctx->i] + ctx->S[ctx->j]) % 256];
}

void
arc4_crypt(struct arc4_ctx *ctx,
           uint8_t *obuf, uint8_t const *ibuf, size_t buflen)
{
  size_t k;

  for (k = 0; k < buflen; k += 1) {
    obuf[k] = ibuf[k] ^ arc4_pad(ctx);
  }
}

void
arc4_crypt_buffer(uint8_t const *key, size_t keylen,
                  uint8_t *buf, size_t buflen)
{
  struct arc4_ctx ctx;

  arc4_init(&ctx, key, keylen);
  arc4_crypt(&ctx, buf, buf, buflen);
}

void
arc4_hash(uint8_t const *buf, size_t buflen,
          uint8_t *hash)
{
  struct arc4_ctx ctx;
  int             i;

  arc4_init(&ctx, buf, buflen);
  for (i = 0; i < ARC4_HASHLEN; i += 1) {
    hash[i] = arc4_pad(&ctx);
  }
}
