#include <unistd.h>

int
main(int argc, char *argv[])
{
  char c;

  while (read(0, &c, 1)) {
    c ^= 3;
    write(1, &c, 1);
  }
  return 0;
}
