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
  char   filename[100];
  int    filenamelen;
  int    fd;
  int    ret;
  time_t now = time(NULL);

  if (! team_exists(teamhash)) {
    return -2;
  }

  linelen = snprintf(line, sizeof(line),
                     "%u %s %s %d\n",
                     now, teamhash, category, points);
  if (sizeof(line) == linelen) {
    return -1;
  }

  /* At one time I had this writing to a single log file, using lockf.
     This works, as long as nobody ever tries to edit the log file.
     Editing the log file would require locking it, which would block
     everything trying to score, effectively taking down the entire
     contest.  If you can't lock it first (nothing in busybox lets you
     do this), you have to bring down pretty much everything manually
     anyway.

     By putting new scores into new files and periodically appending
     those files to the main log file, it is possible to stop the thing
     that appends, edit the file at leisure, and then start the appender
     back up, all without affecting things trying to score: they're
     still able to record their score and move on.  You don't even
     really need an appender, but it does make things look a little
     nicer on the fs.

     The fact that this makes the code simpler is just gravy.
  */

  filenamelen = snprintf(filename, sizeof(filename),
                         "%s/%d.%d.%s.%s.%d",
                         pointsdir, now, getpid(),
                         teamhash, category, points);
  if (sizeof(filename) == filenamelen) {
    return -1;
  }

  fd = open(filename, O_WRONLY | O_CREAT | O_EXCL, 0666);
  if (-1 == fd) {
    return -1;
  }

  write(fd, line, linelen);
  close(fd);
  return 0;
}
