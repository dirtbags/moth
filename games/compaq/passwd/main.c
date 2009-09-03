#include <stdio.h>
#include <string.h>

int
main(int argc, char *argv[])
{
  char reply[256];

  printf("What is the password? ");
  fflush(stdout);
  if (fgets(reply, sizeof(reply), stdin)) {
    char *p;

    for (p = reply; *p; p += 1) {
      if ('\n' == *p) {
        *p = '\0';
        break;
      }
    }
    if (strcmp(reply, passwd)) {
      printf("No it isn't!\n");
    } else {
      printf("Congratulations.\n");
    }
  }

  return 0;
}
