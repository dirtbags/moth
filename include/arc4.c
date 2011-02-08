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
arc4_out(struct arc4_ctx *ctx)
{
  ctx->i = (ctx->i + 1) % 256;
  ctx->j = (ctx->j + ctx->S[ctx->i]) % 256;
  swap(ctx->S[ctx->i], ctx->S[ctx->j]);
  return ctx->S[(ctx->S[ctx->i] + ctx->S[ctx->j]) % 256];
}

void
arc4_crypt(struct arc4_ctx *ctx,
           uint8_t *obuf, const uint8_t *ibuf, size_t buflen)
{
  size_t k;

  for (k = 0; k < buflen; k += 1) {
    obuf[k] = ibuf[k] ^ arc4_out(ctx);
  }
}

void
arc4_crypt_buffer(const uint8_t *key, size_t keylen,
                  uint8_t *buf, size_t buflen)
{
  struct arc4_ctx ctx;

  arc4_init(&ctx, key, keylen);
  arc4_crypt(&ctx, buf, buf, buflen);
}


#ifdef ARC4_MAIN

#include <stdio.h>
#include <sysexits.h>
#include <string.h>

int
main(int argc, char *argv[])
{
  struct arc4_ctx ctx;

  /* Read key and initialize context */
  {
    uint8_t  key[256];
    size_t   keylen = 0;
    char    *ekey   = getenv("KEY");
    FILE    *f;

    if (argc == 2) {
      if (! (f = fopen(argv[1], "r"))) {
        perror(argv[0]);
      }
    } else {
      f = fdopen(3, "r");
    }

    if (f) {
      keylen = fread(key, 1, sizeof(key), f);
      fclose(f);
    } else if (ekey) {
      keylen = strlen(ekey);
      if (keylen > sizeof(key)) {
        keylen = sizeof(key);
      }
      memcpy(key, ekey, keylen);
    }

    if (0 == keylen) {
      fprintf(stderr, "Usage: %s [KEYFILE] <PLAINTEXT\n", argv[0]);
      fprintf(stderr, "\n");
      fprintf(stderr, "You can also pass in the key on fd 3 or in\n");
      fprintf(stderr, "$KEY; omit KEYFILE in this case.\n");
      return EX_IOERR;
    }
    arc4_init(&ctx, key, (size_t)keylen);
  }

  /* Encrypt */
  while (1) {
    int c = getchar();

    if (EOF == c) break;
    putchar(c ^ arc4_out(&ctx));
  }

  return 0;
}

#endif /* ARC4_MAIN */
