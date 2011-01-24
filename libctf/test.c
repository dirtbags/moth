#include <stdio.h>
#include <stdint.h>
#include "rand.h"
#include "md5.h"
#include "token.h"

int
main()
{
  int i;
  uint8_t zeroes[64] = {0};
  uint8_t digest[MD5_DIGEST_LEN];

  for (i = 0; i < 10; i += 1) {
    printf("%d ", randu32() % 10);
  }

  printf("\n4ae71336e44bf9bf79d2752e234818a5\n");

  md5_digest(zeroes, 16, digest);
  for (i = 0; i < sizeof(digest); i += 1) {
    printf("%02x", digest[i]);
  }
  printf("\n");

  {
    char hd[MD5_HEXDIGEST_LEN + 1] = {0};

    md5_hexdigest(zeroes, 16, hd);
    printf("%s\n", hd);
  }

  {
    ssize_t len;
    char    token[TOKEN_MAX];

    len = read_token("foo", 0, 4, token, sizeof(token));
    if (-1 != len) {
      printf("rut roh\n");
    } else {
      printf("Good.\n");
    }
  }

  return 0;
}
