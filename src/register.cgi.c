#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
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

    len = cgi_item(key, sizeof(key));
    if (0 == len) break;
    if ((1 == len) && ('t' == key[0])) {
      teamlen = cgi_item(team, sizeof(team));
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
      cgi_error("The full path to the team hash file is too long.");
    }
    fd = open(filename, 0444, O_WRONLY | O_CREAT | O_EXCL);
    if (-1 == fd) {
      cgi_page("Bad team name",
               ("<p>Either that team name is already in use, or you "
                "found a hash collision (way to go). "
                "In any case, you're going to "
                "have to pick something else.</p>"
                "<p>If you're just trying to find your team hash again,"
                "it's <samp>%s</samp>.</p>"),
               hash);
    }
    write(fd, team, teamlen);
    close(fd);
  }

  /* Let them know what their hash is. */
  cgi_page("Team registered",
            ("<p>Team hash: <samp>%s</samp></p>"
             "<p><b>Save your team hash somewhere!</b>.  You will need it "
             "to claim points.</b></p>"),
            hash);

  return 0;
}
