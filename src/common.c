#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdio.h>
#include <ctype.h>
#include <time.h>
#include "common.h"

/*
 * CGI
 */
static size_t inlen = 0;
static int is_cgi = 0;

int
cgi_init()
{
  char *rm = getenv("REQUEST_METHOD");

  if (! (rm && (0 == strcmp(rm, "POST")))) {
    printf("405 Method not allowed\r\n"
           "Allow: POST\r\n"
           "Content-type: text/html\r\n"
           "\r\n"
           "<h1>Method not allowed</h1>\n"
           "<p>I only speak POST.  Sorry.</p>\n");
    return -1;
  }

  inlen = atoi(getenv("CONTENT_LENGTH"));
  is_cgi = 1;

  return 0;
}

static int
read_char()
{
  if (inlen) {
    inlen -= 1;
    return getchar();
  }
  return EOF;
}

static char
tonum(int c)
{
  if ((c >= '0') && (c <= '9')) {
    return c - '0';
  }
  if ((c >= 'a') && (c <= 'f')) {
    return 10 + c - 'a';
  }
  if ((c >= 'A') && (c <= 'F')) {
    return 10 + c - 'A';
  }
  return 0;
}

static char
read_hex()
{
  int a = read_char();
  int b = read_char();

  return tonum(a)*16 + tonum(b);
}

/* Read a key or a value.  Since & and = aren't supposed to appear
   outside of boundaries, we can use the same function for both.
*/
size_t
cgi_item(char *str, size_t maxlen)
{
  int    c;
  size_t pos = 0;

  while (1) {
    c = read_char();
    switch (c) {
      case EOF:
      case '=':
      case '&':
        str[pos] = '\0';
        return pos;
      case '%':
        c = read_hex();
        break;
      case '+':
        c = ' ';
        break;
    }
    if (pos < maxlen - 1) {
      str[pos] = c;
      pos += 1;
    }
  }
}

void
cgi_head(char *title)
{
  if (is_cgi) {
    printf("Content-type: text/html\r\n\r\n");
  }
  printf(("<!DOCTYPE html>\n"
          "<html>\n"
          "  <head>\n"
          "    <title>%s</title>\n"
          "    <link rel=\"stylesheet\" href=\"ctf.css\" type=\"text/css\">\n"
          "  </head>\n"
          "  <body>\n"
          "    <h1>%s</h1>\n"),
         title, title);
}

void
cgi_foot()
{
  printf("\n"
         "  </body>\n"
         "</html>\n");
}

void
cgi_page(char *title, char *fmt, ...)
{
  va_list  ap;

  cgi_head(title);
  va_start(ap, fmt);
  vprintf(fmt, ap);
  va_end(ap);
  cgi_foot();
  exit(0);
}

void
cgi_error(char *fmt, ...)
{
  va_list ap;

  printf("500 Internal Error\r\n"
         "Content-type: text/plain\r\n"
         "\r\n");
  va_start(ap, fmt);
  vprintf(fmt, ap);
  va_end(ap);
  printf("\n");
  exit(0);
}


/*
 * Common routines
 */


#define EOL(c) ((EOF == (c)) || (0 == (c)) || ('\n' == (c)))

int
fgrepx(char const *needle, char const *filename)
{
  FILE       *f;
  int         found = 0;
  char const *p     = needle;

  f = fopen(filename, "r");
  if (f) {
    while (1) {
      int c = fgetc(f);

      /* This list of cases would have looked so much nicer in OCaml.  I
         apologize. */
      if (EOL(c) && (0 == *p)) {
        found = 1;
        break;
      } else if (EOF == c) {
        break;
      } else if ((0 == p) || (*p != c)) {
        p = needle;
        do {
          c = fgetc(f);
        } while (! EOL(c));
      } else if ('\n' == c) {
        p = needle;
      } else {
        p += 1;
      }
    }
    fclose(f);
  }

  return found;
}

int
my_snprintf(char *buf, size_t buflen, char *fmt, ...)
{
  int     len;
  va_list ap;

  va_start(ap, fmt);
  len = vsnprintf(buf, buflen - 1, fmt, ap);
  va_end(ap);
  if (len >= 0) {
    buf[len] = '\0';
    return len;
  } else {
    return -1;
  }
}

int
team_exists(char const *teamhash)
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
  if (sizeof(filename) <= ret) {
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
award_points(char const *teamhash,
             char const *category,
             long points)
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
                     "%u %s %s %ld\n",
                     now, teamhash, category, points);
  if (sizeof(line) <= linelen) {
    return -1;
  }

  /* At one time I had this writing to a single log file, using lockf.
     This works, as long as nobody ever tries to edit the log file.
     Editing the log file would require locking it, which would block
     everything trying to score, effectively taking down the entire
     contest.  If you can't lock it first--and nothing in busybox lets
     you do this--you have to bring down pretty much everything manually
     anyway.

     By putting new scores into new files and periodically appending
     those files to the main log file, it is possible to stop the thing
     that appends, edit the file at leisure, and then start the appender
     back up, all without affecting things trying to score: they're
     still able to record their score and move on.  You don't even
     really need an appender, but it does make things look a little
     nicer on the fs.

     The fact that this makes the code simpler is just gravy.

     Note that doing this means there's a little time between when a
     score's written and when the scoreboard (which only reads the log
     file) picks it up.  It's not a big deal for the points log, but
     this situation makes this technique unsuitable for writing log
     files that prevent people from double-scoring, like the puzzler or
     token log.
  */

  filenamelen = snprintf(filename, sizeof(filename),
                         "%s/%d.%d.%s.%s.%ld",
                         pointsdir, now, getpid(),
                         teamhash, category, points);
  if (sizeof(filename) <= filenamelen) {
    return -1;
  }

  fd = open(filename, O_WRONLY | O_CREAT | O_EXCL, 0666);
  if (-1 == fd) {
    return -1;
  }

  if (-1 == write(fd, line, linelen)) {
    close(fd);
    return -1;
  }

  close(fd);
  return 0;
}

void
award_and_log_uniquely(char const *team,
                       char const *category,
                       long points,
                       char const *logfile,
                       char const *fmt, ...)
{
  char    line[200];
  int     len;
  int     ret;
  int     fd;
  va_list ap;

  /* Make sure they haven't already claimed these points */
  va_start(ap, fmt);
  len = vsnprintf(line, sizeof(line), fmt, ap);
  va_end(ap);
  if (sizeof(line) <= len) {
    cgi_error("Log line too long");
  }
  if (fgrepx(line, logfile)) {
    cgi_page("Already claimed",
             "<p>Your team has already claimed these points.</p>");
  }

  /* Open and lock logfile */
  fd = open(logfile, O_WRONLY | O_CREAT, 0666);
  if (-1 == fd) {
    cgi_error("Unable to open log");
  }
  if (-1 == lockf(fd, F_LOCK, 0)) {
    cgi_error("Unable to lock log");
  }

  /* Award points */
  if (0 != award_points(team, category, points)) {
    cgi_error("Unable to award points");
  }

  /* Log that we did so */
  /* We can turn that trailing NUL into a newline now since write
     doesn't use C strings */
  line[len] = '\n';
  lseek(fd, 0, SEEK_END);
  if (-1 == write(fd, line, len+1)) {
    cgi_error("Unable to append log");
  }
  close(fd);
}
