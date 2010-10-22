#include <stdio.h>
#include <time.h>
#include <stdint.h>
#include <string.h>
#include "token.h"

/* This hopefully requires an LD_PRELOAD */

uint8_t const key[] = {0x94, 0xf2, 0x92, 0x45,
                       0x12, 0x44, 0x80, 0xe1,
                       0x95, 0x64, 0xcd, 0xe4,
                       0xff, 0x0a, 0x00, 0x10};

int
main(int argc, char *argv[])
{
  char   token[200];
  size_t tokenlen;

  /* Do some bullshit.  Override with:
   *
   * void strcmp(char *a, char *b)
   * {
   *   return 0;
   * }
   */
  {
    FILE         *f = fopen("/dev/urandom", "r");
    unsigned int  seed;
    char          seed_str[50];

    printf("Checking credentials...\n");
    fread(&seed, sizeof(seed), 1, f);
    sprintf(seed_str, "%d", seed);
    if ((argc != 2) || strcmp(seed_str, argv[1])) {
      printf("Ah ah ah!  You didn't say the magic word!\n");
      return 1;
    }
  }

  tokenlen = read_token("ltraceme",
                        key, sizeof(key),
                        token, sizeof(token) - 1);
  if (-1 == tokenlen) {
    write(1, "Something is broken\nI can't read my token.\n", 43);
    return 1;
  }
  token[tokenlen++] = '\0';

  /* You could override this with:
   *
   * void printf(char *fmt, size_t len, char *buf)
   * {
   *   if (fmt[0] == 'T') write(1, buf, len);
   * }
   */
  printf("Token length %u at %p.\n", (unsigned int)tokenlen, token);

  return 0;
}
