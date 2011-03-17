#include <stdio.h>
#include <unistd.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <sysexits.h>
#include "xxtea.h"

#define DUMPf(fmt, args...) fprintf(stderr, "%s:%s:%d " fmt "\n", __FILE__, __FUNCTION__, __LINE__, ##args)
#define DUMP() DUMPf("")
#define DUMP_d(v) DUMPf("%s = %d", #v, v)
#define DUMP_x(v) DUMPf("%s = 0x%x", #v, v)
#define DUMP_s(v) DUMPf("%s = %s", #v, v)
#define DUMP_c(v) DUMPf("%s = '%c' (0x%02x)", #v, v, v)
#define DUMP_p(v) DUMPf("%s = %p", #v, v)

#define min(a,b) (((a)<(b))?(a):(b))

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
  uint8_t *buf     = NULL;
  size_t   len     = 0;
  uint32_t key[4] = {0};

  {
    char *ekey = getenv("KEY");

    if (ekey) {
      memcpy(key, ekey, min(strlen(ekey), sizeof(key)));
    } else {
      read(3, key, sizeof(key));
    }
  }

  while (1) {
    size_t pos = len;
    ssize_t nret;

    buf = realloc(buf, len + 4096);
    if (! buf) {
      perror("realloc");
      return EX_OSERR;
    }

    nret = read(0, buf + pos, 4096);
    if (0 == nret) break;
    if (-1 == nret) {
      perror("read");
      return EX_OSERR;
    }

    len = pos + nret;
  }

  if (argv[1] && (0 == strcmp(argv[1], "-e"))) {
    if (0 == buf[len-1]) {
      fprintf(stderr, "I can't cope with trailing NULs.\n");
      return 1;
    }

    /* Pad out with NUL */
    while (len % 4 > 0) {
      buf[len++] = 0;
    }

    tea_encode(key, (uint32_t *)buf, len/4);
  } else {
    if (len % 4) {
      fprintf(stderr, "Incorrect padding.\n");
      return 1;
    }
    tea_decode(key, (uint32_t *)buf, len/4);

    /* Remove padding.  If your input had trailing NULs, you shouldn't
       use this. */
    while (0 == buf[len-1]) {
      len -= 1;
    }
  }
  write(1, buf, len);

  return 0;
}
