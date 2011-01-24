#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stddef.h>
#include <stdint.h>
#include <time.h>
#include "arc4.h"

/*
 *
 * Random numbers
 *
 */

void
urandom(uint8_t *buf, size_t buflen)
{
  static int             initialized = 0;
  static struct arc4_ctx ctx;

  if (! initialized) {
    int fd = open("/dev/urandom", O_RDONLY);

    if (-1 == fd) {
      struct {
        time_t time;
        pid_t  pid;
      } bits;

      bits.time = time(NULL);
      bits.pid  = getpid();
      arc4_init(&ctx, (uint8_t *)&bits, sizeof(bits));
    } else {
      uint8_t key[256];

      read(fd, key, sizeof(key));
      close(fd);
      arc4_init(&ctx, key, sizeof(key));
    }

    initialized = 1;
  }

  while (buflen--) {
    *(buf++) = arc4_out(&ctx);
  }
}

int32_t
rand32()
{
  int32_t ret;

  urandom((uint8_t *)&ret, sizeof(ret));
  return ret;
}

uint32_t
randu32()
{
  uint32_t ret;

  urandom((uint8_t *)&ret, sizeof(ret));
  return ret;
}
