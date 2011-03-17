#include <stdint.h>
#include <stdlib.h>
#include "arc4.h"

#define DUMPf(fmt, args...) fprintf(stderr, "%s:%s:%d " fmt "\n", __FILE__, __FUNCTION__, __LINE__, ##args)
#define DUMP() DUMPf("")
#define DUMP_d(v) DUMPf("%s = %d", #v, v)
#define DUMP_x(v) DUMPf("%s = 0x%x", #v, v)
#define DUMP_s(v) DUMPf("%s = %s", #v, v)
#define DUMP_c(v) DUMPf("%s = '%c' (0x%02x)", #v, v, v)
#define DUMP_p(v) DUMPf("%s = %p", #v, v)

#define swap(a, b) do {uint8_t _swap=a; a=b, b=_swap;} while (0)

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

/* Create a nonce as an arc4 stream with key=seed */
void
arc4_nonce(uint8_t *nonce, size_t noncelen,
           void *seed, size_t seedlen)
{
  struct arc4_ctx ctx;
  int             i;

  arc4_init(&ctx, seed, seedlen);
  for (i = 0; i < noncelen; i += 1) {
    nonce[i] = arc4_out(&ctx);
  }
}


#ifdef ARC4_MAIN

#include <stdio.h>
#include <sysexits.h>
#include <time.h>
#include <string.h>
#include <sys/types.h>
#include <unistd.h>

int
usage(const char *prog)
{
  fprintf(stderr, "Usage: %s [-e] <PLAINTEXT\n", prog);
  fprintf(stderr, "\n");
  fprintf(stderr, "You must pass in a key on fd 3 or in the environment variable KEY.\n");
  return EX_USAGE;
}

int
main(int argc, char *argv[])
{
  struct arc4_ctx ctx;
  uint8_t         key[ARC4_KEYLEN] = {0};
  size_t          keylen;
  uint8_t         nonce[ARC4_KEYLEN];
  time_t          seed;
  int             i;

  /* Read key and initialize context */
  {
    char *ekey = getenv("KEY");

    if (ekey) {
      keylen = strlen(ekey);
      memcpy(key, ekey, keylen);
    } else {
      FILE *f = fdopen(3, "r");

      if (NULL == f) {
        return usage(argv[0]);
      }

      keylen = fread(key, 1, ARC4_KEYLEN, f);
      fclose(f);
    }
  }

  if (argv[1] && (0 == strcmp(argv[1], "-e"))) {
    seed = time(NULL) * getpid();
    fwrite("arc4", 1, 4, stdout);
    fwrite(&seed, sizeof(seed), 1, stdout);
  } else if (argv[1]) {
    return usage(argv[0]);
  } else {
    char   sig[4];

    fread(&sig, sizeof(sig), 1, stdin);
    if (memcmp(sig, "arc4", 4)) {
      fprintf(stderr, "%s: error: Input is not arc4-encrypted.", argv[0]);
      return 1;
    }
    fread(&seed, sizeof(seed), 1, stdin);
  }

  arc4_nonce(nonce, sizeof(nonce), &seed, sizeof(seed));

  /* Xor key with nonce */
  for (i = 0; i < sizeof(key); i += 1) {
    key[i] ^= nonce[i];
  }

  arc4_init(&ctx, key, sizeof(key));

  while (1) {
    int c = getchar();

    if (EOF == c) break;
    putchar(c ^ arc4_out(&ctx));
  }

  return 0;
}

#endif /* ARC4_MAIN */
