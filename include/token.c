#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stddef.h>
#include <stdlib.h>
#include <unistd.h>
#include <values.h>

#ifndef CTF_BASE
#define CTF_BASE "/var/lib/ctf"
#endif

/*
 *
 * ARC-4 stuff
 *
 */

struct arc4_ctx {
  uint8_t S[256];
  uint8_t i;
  uint8_t j;
};

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

/*
 *
 */



ssize_t
read_token_fd(int fd,
              uint8_t const *key, size_t keylen,
              char *buf, size_t buflen)
{
  ssize_t ret;

  ret = read(fd, buf, buflen);
  if (-1 != ret) {
    arc4_crypt_buffer(key, keylen, (uint8_t *)buf, (size_t)ret);
  }
  return ret;
}


ssize_t
read_token(char const *name,
           uint8_t const *key, size_t keylen,
           char *buf, size_t buflen)
{
  char    path[PATH_MAX];
  int     pathlen;
  int     fd;
  ssize_t ret;

  pathlen = snprintf(path, sizeof(path) - 1,
                     CTF_BASE "/tokens/%s", name);
  path[pathlen] = '\0';

  fd = open(path, O_RDONLY);
  if (-1 == fd) return -1;
  ret = read_token_fd(fd, key, keylen, buf, buflen);
  close(fd);
  return ret;
}
