#include <unistd.h>
#include "token.h"

uint8_t const key[] = {0x5f, 0x64, 0x13, 0x29,
                       0x2e, 0x46, 0x76, 0xcd,
                       0x65, 0xff, 0xe8, 0x03,
                       0xa4, 0xa9, 0x4f, 0xd9};

int
main(int argc, char *argv[])
{
  char    token[200];
  ssize_t tokenlen;

  tokenlen = read_token("gimmie",
                        key, sizeof(key),
                        token, sizeof(token) - 1);
  if (-1 == tokenlen) {
    write(1, "Something is broken\nI can't read my token.\n", 43);
    return 69;
  }

  token[tokenlen++] = '\n';
  write(1, token, tokenlen);

  return 0;
}
