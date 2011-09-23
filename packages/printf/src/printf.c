#include <stdio.h>
#include <ctype.h>
#include <stdlib.h>

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

/* Storage space for tokens */
char *token[5] = {
  "printf:xylep-radar-nanox",
  "printf:xylep-radar-nanox",
  "printf:xylep-radar-nanox",
  "printf:xylep-radar-nanox",
  "printf:xylep-radar-nanox"
};

/* Make this global so the stack isn't gigantic */
char global_fmt[8000] = {0};


int
main(int argc, char *argv[], char *env[])
{
  char *t0          = token[0];
  int   t1[100];
  char *fmt         = global_fmt;
  char *datacomp    = "welcome datacomp";
  int   token4_flag = 0;
  int   i;

  /* Make stderr buffer until lines */
  setlinebuf(stderr);

  /* So the compiler won't complain about unused variables */
  i = datacomp[0] ^ t0[0];

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
