#include <stdio.h>
#include <sysexits.h>
#include "token.h"

uint8_t const key[] = {0x5f, 0x64, 0x13, 0x29,
                       0x2e, 0x46, 0x76, 0xcd,
                       0x65, 0xff, 0xe8, 0x03,
                       0xa4, 0xa9, 0x4f, 0xd9};

int
main(int argc, char *argv[])
{
  if (-1 == print_token("gimmie", key, sizeof(key))) {
    fprintf(stderr, "Something is broken; I can't read my token.\n");
    return EX_UNAVAILABLE;
  }

  return 0;
}
