#include <stdio.h>
#include <sysexits.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  if (2 != argc) {
    fprintf(stderr, "Usage: %s CATEGORY\n", argv[0]);
    return EX_USAGE;
  }

  /* Create the token. */
  {
    unsigned char crap[itokenlen];
    unsigned char digest[bubblebabble_len(itokenlen)];

    urandom((char *)crap, sizeof(crap));

    /* Digest some random junk. */
    bubblebabble(digest, (unsigned char *)&crap, itokenlen);

    /* Append digest to category name. */
    printf("%s:%s\n", argv[1], digest);
  }

  return 0;
}
