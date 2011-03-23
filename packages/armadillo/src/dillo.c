#include <unistd.h>
#include <time.h>
#include <stdint.h>
#include "arc4.h"
#include "token.h"

const uint8_t key[] =
  {0xa5, 0xb1, 0x6f, 0xce,
   0x59, 0x2d, 0xb1, 0xe9,
   0x4b, 0x07, 0x91, 0x6d,
   0x9f, 0x3b, 0xc8, 0xc6};

const char dillo[] = 
  ("           .::7777::-.\n"
   "          /:'////' `::>/|/\n"
   "        .',  ||||   `/( e\\\n"
   "    -==~-'`-Xm````-mr' `-_\\\n");

int
main(int argc, char *argv[])
{
  uint8_t v;
  int     i;

  /* Pick a random non-zero xor value */
  do {
    v = arc4_rand8();
  } while (! v);


  /* Print the dillo */
  for (i = 0; dillo[i]; i += 1) {
    struct timespec req = {0, 33000000};
    uint8_t         c   = dillo[i];

    if ('\n' != c) {
      c ^= v;
    }
    write(1, &c, 1);
    nanosleep(&req, NULL);
  }

  /* Read a single byte; strace will help with solution */
  {
    uint8_t c;

    read(0, &c, 1);
    if (c != v) {
      return 1;
    }
  }

  if (-1 == print_token("dillo", key, sizeof(key))) {
    write(2, "Something is broken; I can't read my token.\n", 44);
    return 69;
  }

  return 0;
}
