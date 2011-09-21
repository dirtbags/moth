#define _GNU_SOURCE
#include <stdio.h>
#include <stdint.h>
#include <limits.h>
#include "token.h"
#include "arc4.h"

#ifndef CTF_BASE
#define CTF_BASE "/var/lib/ctf"
#endif

ssize_t
write_token(FILE *out,
            const char *name,
            const uint8_t *key, size_t keylen)
{
  char    *base;
  char     path[PATH_MAX];
  int      pathlen;
  FILE    *f;
  ssize_t  ret;

  base = getenv("CTF_BASE");
  if (! base) base = CTF_BASE;

  pathlen = snprintf(path, sizeof(path) - 1,
                     "%s/tokens/%s", base, name);
  path[pathlen] = '\0';

  f = fopen(path, "r");
  if (NULL == f) return -1;
  ret = arc4_decrypt_stream(out, f, key, keylen);
  fclose(f);

  return ret;
}

ssize_t
print_token(const char *name,
            const uint8_t *key, size_t keylen)
{
  return write_token(stdout, name, key, keylen);
}

ssize_t
get_token(char *buf, size_t buflen,
          const char *name,
          const uint8_t *key, size_t keylen)
{
  FILE    *f;
  ssize_t  l;

  f = fmemopen(buf, buflen, "w");
  if (! f) return -1;

  l = write_token(f, name, key, keylen);
  fclose(f);
  if (-1 == l) return -1;

  while ('\n' == buf[l-1]) l -= 1;
  buf[l-1] = '\0';

  return l;
}
