#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <sysexits.h>
#include "arc4.h"
#include "token.h"

#define ROUNDS 20

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

  for (i = 1; i < 8; i += 1) {
    signal(i, handler);
  }

  for (i = 0; i < ROUNDS; i += 1) {
    int desired = (arc4_rand8() % 7) + 1;

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

  if (-1 == print_token("killme", key, sizeof(key))) {
    fprintf(stderr, "Something is broken; I can't read my token.\n");
    return EX_UNAVAILABLE;
  }

  return 0;
}
