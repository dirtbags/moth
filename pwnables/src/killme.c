#include <signal.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <time.h>
#include "token.h"

#define SIGS 20

uint8_t const key[] = {0x51, 0x91, 0x6d, 0x81,
                       0x14, 0x21, 0xf8, 0x95,
                       0xb8, 0x09, 0x87, 0xa6,
                       0xa8, 0xb0, 0xa0, 0x46};

int lastsig;

void
handler(int signum)
{
  lastsig = signum;
}

int
main(int argc, char *argv[])
{
  int i;

  {
    /* Seed random number generator */
    FILE *f;
    int seed;

    f = fopen("/dev/urandom", "r");
    if (f) {
      fread(&seed, sizeof(seed), 1, f);
      srandom(seed);
    } else {
      srandom(getpid() * time(NULL));
    }
  }

  for (i = 1; i < 8; i += 1) {
    signal(i, handler);
  }

  for (i = 0; i < SIGS; i += 1) {
    int desired = (random() % 7) + 1;

    lastsig = 0;
    printf("%d\n", desired);
    fflush(stdout);
    if (i == 0) {
      sleep(5);
    } else {
      sleep(1);
    }
    if (0 == lastsig) {
      printf("Too slow.\n");
      return 1;
    }
    if (lastsig != desired) {
      printf("Wrong one.\n");
      return 1;
    }
  }

  {
    char   token[200];
    size_t tokenlen;

    tokenlen = read_token("killme",
                          key, sizeof(key),
                          token, sizeof(token) - 1);
  }

  return 0;
}
