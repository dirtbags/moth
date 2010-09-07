#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdio.h>
#include <ctype.h>
#include <time.h>
#include "common.h"

int
team_exists(char *teamhash)
{
  struct stat buf;
  char        filename[100];
  int         ret;
  int         i;

  /* Check for invalid characters. */
  for (i = 0; teamhash[i]; i += 1) {
    if (! isalnum(teamhash[i])) {
      return 0;
    }
  }

  /* Build filename. */
  ret = snprintf(filename, sizeof(filename),
                 "%s/%s", teamdir, teamhash);
  if (sizeof(filename) == ret) {
    return 0;
  }

  /* lstat seems to be the preferred way to check for existence. */
  ret = lstat(filename, &buf);
  if (-1 == ret) {
    return 0;
  }

  return 1;
}

int
award_points(char *teamhash, char *category, int points)
{
  char   line[100];
  int    linelen;
  int    fd;
  int    ret;

  if (! team_exists(teamhash)) {
    return -2;
  }

  linelen = snprintf(line, sizeof(line),
                     "%u %s %s %d\n",
                     time(NULL), teamhash, category, points);
  if (sizeof(line) == linelen) {
    return -1;
  }

  fd = open(pointslog, O_WRONLY | O_CREAT, 0666);
  if (-1 == fd) {
    return -1;
  }

  ret = lockf(fd, F_LOCK, 0);
  if (-1 == ret) {
    close(fd);
    return -1;
  }

  ret = lseek(fd, 0, SEEK_END);
  if (-1 == ret) {
    close(fd);
    return -1;
  }

  write(fd, line, linelen);
  close(fd);
  return 0;
}
