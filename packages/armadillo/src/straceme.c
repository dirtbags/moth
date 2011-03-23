#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <stdarg.h>
#include <sysexits.h>
#include <stdio.h>
#include <ctype.h>
#include "token.h"

uint8_t const key[] = {0x30, 0x00, 0x55, 0x0f,
                       0xc2, 0xf6, 0x52, 0x2a,
                       0x31, 0xfd, 0x00, 0x92,
                       0x9d, 0x49, 0x24, 0xce};

int
main(int argc, char *argv[])
{
  /* Check argv[1].
   *
   * If no args, argv[1] will be NULL, which causes a segfault.
   * This is what we want.
   *
   * To pass this, run ./straceme $$.
   * But you have to do it from a shell script, because if you run
   *    strace ./straceme $$
   * getppid() will return the PID of strace!
   */
  if (getppid() != atoi(argv[1])) {
    write(2, "argv[1] incorrect\n", 18);
    return EX_USAGE;
  }

  /* Read an rc file.
   *
   * To pass this, set $HOME to someplace you have access.
   */
  {
    int  fd;
    char fn[128];
    char bs[1000];
    int  len;

    len = snprintf(fn, sizeof(fn) - 1, "%s/.stracemerc", getenv("HOME"));
    if (len < 0) {
      len = 0;
    }
    fn[len] = '\0';

    if (-1 == (fd = open(fn, O_RDONLY))) {
      fd = open("/etc/stracemerc", O_RDONLY);
    }
    if (-1 == fd) {
      return EX_NOINPUT;
    }

    /* We don't actually care about contents */
    read(fd, bs, sizeof(bs));
    close(fd);
  }

  /* Read in category name from fd 5
   *
   * echo -n straceme > foo.txt
   * ./straceme $$ 5< foo.txt
   */
  {
    char   cat[50];
    int    catlen;
    char   token[200];
    size_t tokenlen;
    int    i;

    catlen = read(5, cat, sizeof(cat) - 1);
    for (i = 0; i < catlen; i += 1) {
      if (! isalnum(cat[i])) break;
    }
    cat[i] = '\0';

    if (-1 == print_token(cat, key, sizeof(key))) {
      write(2, "Something is broken; I can't read my token.\n", 44);
      return 69;
    }
  }
  return 0;
}
