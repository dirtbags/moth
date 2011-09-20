#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "arc4.h"

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


/***************************************************
 *
 * Psuedo Random Number Generation
 *
 */
static struct arc4_ctx prng_ctx;
static int             prng_initialized = 0;

void
arc4_rand_seed(const uint8_t *seed, size_t seedlen)
{
  arc4_init(&prng_ctx, seed, seedlen);
  prng_initialized = 1;
}

static void
arc4_rand_autoseed()
{
  if (! prng_initialized) {
    uint8_t  key[ARC4_KEYLEN];
    FILE    *urandom;

    /* Open /dev/urandom or die trying */
    urandom = fopen("/dev/urandom", "r");
    if (! urandom) {
      perror("Opening /dev/urandom");
      abort();
    }
    setbuf(urandom, NULL);
    fread(&key, sizeof(key), 1, urandom);
    fclose(urandom);

    arc4_rand_seed(key, sizeof(key));
  }
}

uint8_t
arc4_rand8()
{
  arc4_rand_autoseed();
  return arc4_out(&prng_ctx);
}

uint32_t
arc4_rand32()
{
  arc4_rand_autoseed();
  return ((arc4_out(&prng_ctx) << 0) |
          (arc4_out(&prng_ctx) << 8) |
          (arc4_out(&prng_ctx) << 16) |
          (arc4_out(&prng_ctx) << 24));
}

/*****************************************
 *
 * Stream operations
 *
 */

ssize_t
arc4_encrypt_stream(FILE *out, FILE *in,
                    const uint8_t *key, size_t keylen)
{
  struct arc4_ctx ctx;
  uint32_t        seed    = arc4_rand32();
  uint8_t         nonce[ARC4_KEYLEN];
  ssize_t         written = 0;
  int             i;

  fwrite("arc4", 4, 1, out);
  fwrite(&seed, sizeof(seed), 1, out);

  arc4_nonce(nonce, sizeof(nonce), &seed, sizeof(seed));
  for (i = 0; i < keylen; i += 1) {
    nonce[i] ^= key[i];
  }
  arc4_init(&ctx, nonce, sizeof(nonce));

  while (1) {
    int c = fgetc(in);

    if (EOF == c) break;
    fputc((uint8_t)c ^ arc4_out(&ctx), out);
    written += 1;
  }

  return written;
}

int
arc4_decrypt_stream(FILE *out, FILE *in,
                    const uint8_t *key, size_t keylen)
{
  struct arc4_ctx ctx;
  uint32_t        seed;
  uint8_t         nonce[ARC4_KEYLEN];
  ssize_t         written = 0;
  char            sig[4];
  int             i;

  fread(&sig, sizeof(sig), 1, in);
  if (memcmp(sig, "arc4", 4)) {
    return -1;
  }
  fread(&seed, sizeof(seed), 1, in);

  arc4_nonce(nonce, sizeof(nonce), &seed, sizeof(seed));
  for (i = 0; i < keylen; i += 1) {
    nonce[i] ^= key[i];
  }
  arc4_init(&ctx, nonce, sizeof(nonce));

  while (1) {
    int c = fgetc(in);

    if (EOF == c) break;
    fputc((uint8_t)c ^ arc4_out(&ctx), out);
    written += 1;
  }

  return written;
}


#ifdef ARC4_MAIN

#include <sysexits.h>
#include <unistd.h>

int
main(int argc, char *argv[])
{
  uint8_t key[ARC4_KEYLEN] = {0};
  size_t  keylen;

  /* Read key and initialize context */
  {
    char *ekey = getenv("KEY");

    if (ekey) {
      keylen = strlen(ekey);
      memcpy(key, ekey, keylen);
    } else {
      keylen = read(3, key, sizeof(key));
      if (-1 == keylen) {
        fprintf(stderr, "error: must specify key.\n");
        return 1;
      }
    }
  }

  if (! argv[1]) {
    if (-1 == arc4_decrypt_stream(stdout, stdin, key, keylen)) {
      fprintf(stderr, "error: not an arc4 stream.\n");
      return 1;
    }
  } else if (0 == strcmp(argv[1], "-e")) {
    arc4_encrypt_stream(stdout, stdin, key, keylen);
  } else {
    fprintf(stderr, "Usage: %s [-e] <PLAINTEXT\n", argv[0]);
    fprintf(stderr, "\n");
    fprintf(stderr, "You must pass in a key on fd 3 or in the environment variable KEY.\n");
    return EX_USAGE;
  }

  return 0;
}

#endif /* ARC4_MAIN */
