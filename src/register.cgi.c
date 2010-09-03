#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdarg.h>
#include <ctype.h>
#include "cgi.h"

char *BASE_PATH = "/var/lib/ctf/teams";


unsigned int
djbhash(char const *buf, size_t buflen)
{
  unsigned int h = 5381;

  while (buflen--) {
    h = ((h << 5) + h) ^ *(buf++);
  }
  return h;
}

void
page(char *title, char *fmt, ...)
{
  FILE    *p;
  va_list  ap;

  printf("Content-type: text/html\r\n\r\n");
  fflush(stdout);
  p = popen("./template", "w");
  if (NULL == p) {
    printf("<h1>%s</h1>\n", title);
    p = stdout;
  } else {
    fprintf(p, "Title: %s\n", title);
  }
  va_start(ap, fmt);
  vfprintf(p, fmt, ap);
  va_end(ap);
  fclose(p);
}

int
main(int argc, char *argv[])
{
  char   team[80];
  size_t teamlen;
  char   hash[9];

  if (-1 == cgi_init()) {
    return 0;
  }

  /* Read in team name, the only thing we care about */
  while (1) {
    size_t len;
    char   key[20];

    len = read_item(key, sizeof(key));
    if (0 == len) break;
    if ((1 == len) && ('t' == key[0])) {
      teamlen = read_item(team, sizeof(team));
    }
  }

  /* Compute the hash */
  sprintf(hash, "%08x", djbhash(team, teamlen));

  /* Write team name into file */
  {
    char filename[100];
    int  fd;
    int  ret;

    ret = snprintf(filename, sizeof(filename),
                   "%s/%s",
                   BASE_PATH, hash);
    if (sizeof(filename) == ret) {
      printf(("500 Server screwed up\n"
              "Content-type: text/plain\n"
              "\n"
              "The full path to the team hash file is too long.\n"));
      return 0;
    }
    fd = open(filename, 0444, O_WRONLY | O_CREAT | O_EXCL);
    if (-1 == fd) {
      page("Bad team name",
           ("<p>Either that team name is already in use, or you "
            "found a hash collision (way to go). "
            "In any case, you're going to "
            "have to pick something else.</p>"
            "<p>If you're just trying to find your team hash again,"
            "it's <samp>%s</samp>.</p>"),
           hash);
      return 0;
    }
    write(fd, team, teamlen);
    close(fd);
  }

  /* Let them know what their hash is. */
  page("Team registered",
       ("<p>Team hash: <samp>%s</samp></p>"
        "<p><b>Save your team hash somewhere!</b>.  You will need it "
        "to claim points.</b></p>"),
       hash);

  return 0;
}
