#include <stdio.h>
#include <ctype.h>
#include <stdlib.h>
#include "token.h"

void
record(char *buf) {
  char *p;
  char *ip = getenv("TCPREMOTEIP");

  fprintf(stderr, "%s: ", ip);
  for (p = buf; *p; p += 1) {
    if (isprint(*p)) {
      fputc(*p, stderr);
    } else {
      fprintf(stderr, "%%%02x", *p);
    }
  }
  fputc('\n', stderr);
}

uint8_t const key[] = {0x98, 0x37, 0x92, 0x7d,
                       0xa5, 0x6d, 0xc9, 0x61,
                       0xca, 0x97, 0xf8, 0xa5,
                       0xfe, 0x0f, 0xf6, 0xfc};

#define NTOKENS 5

/* Storage space for tokens */
char token[NTOKENS][TOKEN_MAX];

/* Make this global so the stack isn't gigantic */
char global_fmt[8000] = {0};


/* Since this runs in a chroot jail, and setting up all the symlinks is
 * a pain in the butt, we just read from file discriptors passed in.
 * Pipes are the best thing.  :D
 */
void
read_tokens()
{
  int     i;
  ssize_t len;

  for (i = 0; i < NTOKENS; i += 1) {
    len = read_token_fd(i + 3, key, sizeof(key), token[i], sizeof(token[i]));
    if (len >= sizeof(token[i])) abort();
    token[i][len] = '\0';
    printf("Token %d: %s\n", i, token[i]);
  }
}

int
main(int argc, char *argv[], char *env[])
{
  char *t0          = token[0];
  int   t1[TOKEN_MAX];
  char *fmt         = global_fmt;
  char *datacomp    = "welcome datacomp";
  int   token4_flag = 0;
  int   i;

  /* Make stderr buffer until lines */
  setlinebuf(stderr);

  /* So the compiler won't complain about unused variables */
  i = datacomp[0] ^ t0[0];

  read_tokens();

  /* Token 0 just hangs out on the stack */

  /* Set up token 1 (%c%c%c%c...) */
  for (i = 0; '\0' != token[1][i]; i += 1) {
    t1[i] = token[1][i];
  }
  t1[i-1] = '\n';

  /* Stick token 2 into the environment */
  for (i = 0; env[i]; i += 1);
  env[i-1] = token[2];

  /* token 3 is pretty much a gimmie */

  /* token 4 will only be printed if you set token4_flag to non-zero */

  if (NULL == fgets(global_fmt, sizeof(global_fmt), stdin)) {
    return 0;
  }

  record(fmt);

  printf(fmt,
         "Welcome to the printf category.\n",
         "There are multiple tokens hiding here.\n",
         "Good luck!\n",
         token[3],
         "token4_flag (@ ", &token4_flag, "): ", token4_flag, "\n");
  if (token4_flag) {
    printf("%s\n", token[4]);
  }

  return 0;
}
