#include <stdio.h>
#include <string.h>

char ps1[] = "What is the password? ";
char passwd[] = {0x34, 0x07, 0x05, 0x13,
                 0x45, 0x1b, 0x53, 0x4c,
                 0x15, 0x1d, 0x0b, 0x43,
                 0x18, 0x04, 0x17, 0x53,
                 0x1d, 0x0a, 0x06, 0x00};

int
main(int argc, char *argv[])
{
  char reply[256];

  printf(ps1);
  fflush(stdout);
  if (fgets(reply, sizeof(reply), stdin)) {
    int i;

    for (i = 0; reply[i]; i += 1) {
      if ('\n' == reply[i]) {
        reply[i] = '\0';
        break;
      }
      reply[i] ^= ps1[i % sizeof(ps1)];
      //printf("0x%02x, ", reply[i]);
    }
    if (strcmp(reply, passwd)) {
      printf("No it isn't!\n");
    } else {
      printf("Congratulations.\n");
    }
  }

  return 0;
}
