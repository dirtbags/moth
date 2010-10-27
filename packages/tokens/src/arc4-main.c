#include <stdint.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <sysexits.h>
#include "arc4.h"

int
main(int argc, char *argv[])
{
  struct arc4_ctx ctx;

  /* Read key and initialize context */
  {
    uint8_t  key[256];
    size_t   keylen = 0;
    char    *ekey   = getenv("KEY");
    FILE    *f;

    if (argc == 2) {
      if (! (f = fopen(argv[1], "r"))) {
        perror(argv[0]);
      }
    } else {
      f = fdopen(3, "r");
    }

    if (f) {
      keylen = fread(key, 1, sizeof(key), f);
      fclose(f);
    } else if (ekey) {
      keylen = strlen(ekey);
      if (keylen > sizeof(key)) {
        keylen = sizeof(key);
      }
      memcpy(key, ekey, keylen);
    }

    if (0 == keylen) {
      fprintf(stderr, "Usage: %s [KEYFILE] <PLAINTEXT\n", argv[0]);
      fprintf(stderr, "\n");
      fprintf(stderr, "You can also pass in the key on fd 3 or in\n");
      fprintf(stderr, "$KEY; omit KEYFILE in this case.\n");
      return EX_IOERR;
    }
    arc4_init(&ctx, key, (size_t)keylen);
  }

  /* Encrypt */
  while (1) {
    int c = getchar();

    if (EOF == c) break;
    putchar(c ^ arc4_pad(&ctx));
  }

  return 0;
}
