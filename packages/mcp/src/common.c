#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <values.h>
#include <time.h>
#include "common.h"

#ifdef NODUMP
#  define DUMPf(fmt, args...)
#else
#  define DUMPf(fmt, args...) fprintf(stderr, "%s:%s:%d " fmt "\n", __FILE__, __FUNCTION__, __LINE__, ##args)
#endif
#define DUMP() DUMPf("")
#define DUMP_d(v) DUMPf("%s = %d", #v, v)
#define DUMP_x(v) DUMPf("%s = 0x%x", #v, v)
#define DUMP_s(v) DUMPf("%s = %s", #v, v)
#define DUMP_c(v) DUMPf("%s = '%c' (0x%02x)", #v, v, v)
#define DUMP_p(v) DUMPf("%s = %p", #v, v)


#define POST_MAX 1024

/*
 * CGI
 */
static int is_cgi  = 0;
static char **argv = NULL;

static int
read_char_argv()
{
  static int   arg = 0;
  static char *p;

  if (NULL == argv) {
    return EOF;
  }

  if (0 == arg) {
    arg = 1;
    p = argv[1];
  }

  if (! p) {
    return EOF;
  } else if (! *p) {
    arg += 1;
    p = argv[arg];
    return '&';
  }

  return *(p++);
}

static int
read_char_stdin()
{
  static int inlen = -1;

  if (-1 == inlen) {
    char *p = getenv("CONTENT_LENGTH");
    if (p) {
      inlen = atoi(p);
      if (inlen > POST_MAX) {
        inlen = POST_MAX;
      }
    } else {
      inlen = 0;
    }
  }

  if (inlen) {
    inlen -= 1;
    return getchar();
  }
  return EOF;
}

static int
read_char_query_string()
{
  static char *p = (char *)-1;

  if ((char *)-1 == p) {
    p = getenv("QUERY_STRING");
  }

  if (! p) {
    return EOF;
  } else if (! *p) {
    return EOF;
  } else {
    return *(p++);
  }
}

static int (* read_char)() = read_char_argv;

int
cgi_init(char *global_argv[])
{
  char *rm = getenv("REQUEST_METHOD");

  if (! rm) {
    read_char = read_char_argv;
    argv = global_argv;
  } else if (0 == strcmp(rm, "POST")) {
    read_char = read_char_stdin;
    is_cgi = 1;
  } else if (0 == strcmp(rm, "GET")) {
    read_char = read_char_query_string;
    is_cgi = 1;
  } else {
    printf(("405 Method not allowed\r\n"
            "Allow: GET, POST\r\n"
            "Content-type: text/plain\r\n"
            "\r\n"
            "%s is not allowed.\n"),
           rm);
    return -1;
  }

  return 0;
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


#define EOL(c) ((EOF == (c)) || ('\n' == (c)))

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
      if (EOL(c) && ('\0' == *p)) {
        found = 1;
        break;
      } else if (EOF == c) {    /* End of file */
        break;
      } else if (('\0' == p) || (*p != c)) {
        p = needle;
        /* Discard the rest of the line */
        do {
          c = fgetc(f);
        } while (! EOL(c));
      } else if (EOL(c)) {
        p = needle;
      } else {                  /* It matched */
        p += 1;
      }
    }
    fclose(f);
  }

  return found;
}

void
urandom(char *buf, size_t buflen)
{
  static int fd = -2;

  if (-2 == fd) {
    srandom(time(NULL) * getpid());
    fd = open("/dev/urandom", O_RDONLY);
  }
  if (-1 != fd) {
    int len;

    len = read(fd, buf, buflen);
    if (len == buflen) {
      return;
    }
  }

  /* Fall back to libc's crappy thing */
  {
    int i;

    for (i = 0; i < buflen; i += 1) {
      buf[i] = (char)random();
    }
  }
}

int
my_snprintf(char *buf, size_t buflen, char *fmt, ...)
{
  int     len;
  va_list ap;

  va_start(ap, fmt);
  len = vsnprintf(buf, buflen - 1, fmt, ap);
  va_end(ap);
  buf[buflen - 1] = '\0';
  if (len >= buflen) {
    return buflen - 1;
  } else {
    return len;
  }
}

static char *
mkpath(char const *base, char const *fmt, va_list ap)
{
  char         relpath[PATH_MAX];
  static char  path[PATH_MAX];
  char const  *var;
  int          len;

  len = vsnprintf(relpath, sizeof(relpath) - 1, fmt, ap);
  relpath[sizeof(relpath) - 1] = '\0';

  var = getenv("CTF_BASE");
  if (! var) {
    var = base;
  }

  my_snprintf(path, sizeof(path), "%s/%s", var, relpath);
  return path;
}

char *
state_path(char const *fmt, ...)
{
  va_list  ap;
  char    *ret;

  va_start(ap, fmt);
  ret = mkpath("/var/lib/ctf", fmt, ap);
  va_end(ap);
  return ret;
}

char *
package_path(char const *fmt, ...)
{
  va_list  ap;
  char    *ret;

  va_start(ap, fmt);
  ret = mkpath("/opt", fmt, ap);
  va_end(ap);
  return ret;
}


int
team_exists(char const *teamhash)
{
  struct stat buf;
  int         ret;
  int         i;

  if ((! teamhash) || (! *teamhash)) {
    return 0;
  }

  /* Check for invalid characters. */
  for (i = 0; teamhash[i]; i += 1) {
    if (! isalnum(teamhash[i])) {
      return 0;
    }
  }

  /* stat seems to be the preferred way to check for existence. */
  ret = stat(state_path("teams/names/%s", teamhash), &buf);
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
  char   *filename;
  int    fd;
  time_t now = time(NULL);

  if (! team_exists(teamhash)) {
    return -2;
  }

  linelen = snprintf(line, sizeof(line),
                     "%lu %s %s %ld\n",
                     (unsigned long)now, teamhash, category, points);
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

  filename = state_path("points.tmp/%d.%d.%s.%s.%ld",
                        now, getpid(),
                        teamhash, category, points);

  fd = open(filename, O_WRONLY | O_CREAT | O_EXCL, 0666);
  if (-1 == fd) {
    return -1;
  }

  if (-1 == write(fd, line, linelen)) {
    close(fd);
    return -1;
  }

  close(fd);

  /* Rename into points.new */
  {
    char ofn[PATH_MAX];

    strncpy(ofn, filename, sizeof(ofn));
    filename = state_path("points.new/%d.%d.%s.%s.%ld",
                          now, getpid(),
                          teamhash, category, points);
    rename(ofn, filename);
  }

  return 0;
}

/** Award points iff they haven't been logged.

    If [line] is not in [dbfile], append it and give [points] to [team]
    in [category].
*/
void
award_and_log_uniquely(char const *team,
                       char const *category,
                       long points,
                       char const *dbpath,
                       char const *line)
{
  int   fd;

  /* Make sure they haven't already claimed these points */
  if (fgrepx(line, dbpath)) {
    cgi_page("Already claimed",
             "<p>Your team has already claimed these points.</p>");
  }

  /* Open and lock logfile */
  fd = open(dbpath, O_WRONLY | O_CREAT, 0666);
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
  lseek(fd, 0, SEEK_END);
  if (-1 == write(fd, line, strlen(line))) {
    cgi_error("Unable to append log");
  }
  if (-1 == write(fd, "\n", 1)) {
    cgi_error("Unable to append log");
  }
  close(fd);
}
