#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>
#include <time.h>
#include <stdio.h>

#define XDEBUG

int
once()
{
  char sdrow[25][10];
  int  nwords = 5 + (rand() % 2);
  int  i;

#ifdef DEBUG
  nwords = 2;
#endif

  for (i = 0; i < nwords; i += 1) {
    char *drow = sdrow[i];
    int   len  = 4 + (rand() % 6);
    int   j;

    if (i > 0) putchar(' ');
    for (j = 0; j < len; j += 1) {
      char c = 'a' + (rand() % 26);

      putchar(c);
      drow[len-j-1] = c;
    }

    drow[j] = 0;
  }

#ifdef DEBUG
  printf ("    (answer: ");
  for (i = 0; i < nwords; i += 1) {
    if (i > 0) putchar(' ');
    printf("%s", sdrow[i]);
  }
  putchar(')');
#endif

  putchar('\n');
  fflush(stdout);

  for (i = 0; i < nwords; i += 1) {
    char *p;

    if (i > 0) {
      if (getchar() != ' ') return -1;
    }
    for (p = sdrow[i]; *p; p += 1) {
      int c = getchar();

      if (c != *p) return -1;
    }
  }
  if (getchar() != '\n') return -1;

  return 0;
}


int
main(int argc, char *argv[])
{
  char   token[100];
  int    i;

  {
    FILE *tokenin = fdopen(3, "r");

    if (! tokenin) {
      fprintf(stderr, "Somebody didn't read the instructions.\n");
      return 1;
    }

    if (NULL == fgets(token, sizeof(token), tokenin)) {
      fprintf(stderr, "Error reading token.\n");
      return 1;
    }

    fclose(tokenin);
  }


#ifndef DEBUG
  /* don't hang around forever waiting for input */
  alarm(3);
#endif

  srandom(time(NULL) * getpid());

  for (i = 0; i < 12; i += 1) {
    if (-1 == once()) {
      printf("tahT saw ton tahw I saw gnitcepxe\n");
      return 1;
    }
  }
  fputs(token, stdout);

  return 0;
}
