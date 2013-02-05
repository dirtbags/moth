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
      if (inlen < 0) {
        inlen = 0;
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
cgi_result(int code, char *desc, char *fmt, ...)
{
  va_list ap;

  if (is_cgi) {
    printf("Status: %d %s\r\n", code, desc);
  }
  cgi_head(desc);
  va_start(ap, fmt);
  vprintf(fmt, ap);
  va_end(ap);
  cgi_foot();
  exit(0);
}

void
cgi_fail(int err)
{
    switch (err) {
        case ERR_GENERAL:
            cgi_result(500, "Points not awarded", "<p>The server is unable to award your points at this time.</p>");
        case ERR_NOTEAM:
            cgi_result(409, "No such team", "<p>There is no team with that hash.</p>");
        case ERR_CLAIMED:
            cgi_result(409, "Already claimed", "<p>That is the correct answer, but your team has already claimed these points.</p>");
        default:
            cgi_result(409, "Failure", "<p>Failure code: %d</p>", err);
    }
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
cgi_error(char *text)
{
  cgi_result(500, "Internal error", "<p>%s</p>", text);
}


/*
 * Common routines
 */

/* cut -d$ANCHOR -f2- | grep -Fx "$NEEDLE" */
int
anchored_search(char const *filename, char const *needle, char const anchor)
{
    FILE *f = fopen(filename, "r");
    size_t nlen = strlen(needle);
    char line[1024];
    int ret = 0;

    while (f) {
        char *p;

        if (NULL == fgets(line, sizeof line, f)) {
            break;
        }

        /* Find anchor */
        if (anchor) {
            p = strchr(line, anchor);
            if (! p) {
                continue;
            }
            p += 1;
        } else {
            p = line;
        }

        /* Don't bother with strcmp if string lengths differ.
           If this string is shorter than the previous, it's okay.  This is
           just a performance hack.
         */
        if ((p[nlen] != '\n') &&
            (p[nlen] != '\0')) {
            continue;
        }
        p[nlen] = 0;

        /* Okay, now we can compare! */
        if (0 == strcmp(p, needle)) {
            ret = 1;
            break;
        }
    }

    if (f) {
        fclose(f);
    }

    return ret;
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

void
ctf_chdir()
{
    static int initialized = 0;
    int i;

    if (initialized) {
        return;
    }
    initialized = 1;

    /* chdir to $CTF_BASE */
    {
        char const *ctf_base = getenv("CTF_BASE");

        if (ctf_base) {
            chdir(ctf_base);
        }
    }

    /* Keep going up one directory until there's a packages directory */
    for (i = 0; i < 5; i += 1) {
        struct stat st;

        if ((0 == stat("packages", &st)) &&
               S_ISDIR(st.st_mode)) {
            return;
        }
        chdir("..");
    }
    fprintf(stderr, "Can not determine CTF_BASE directory: exiting.\n");
    exit(66);
}


static char *
mkpath(char const *type, char const *fmt, va_list ap)
{
  char         relpath[PATH_MAX];
  static char  path[PATH_MAX];

  ctf_chdir();
  vsnprintf(relpath, sizeof(relpath) - 1, fmt, ap);
  relpath[sizeof(relpath) - 1] = '\0';

  /* $CTF_BASE/type/relpath */
  my_snprintf(path, sizeof(path), "%s/%s", type, relpath);
  return path;
}

char *
state_path(char const *fmt, ...)
{
  va_list  ap;
  char    *ret;

  va_start(ap, fmt);
  ret = mkpath("state", fmt, ap);
  va_end(ap);
  return ret;
}

char *
package_path(char const *fmt, ...)
{
  va_list  ap;
  char    *ret;

  va_start(ap, fmt);
  ret = mkpath("packages", fmt, ap);
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

/* Return values:
    -1: general error
    -2: no such team
    -3: points already awarded
 */
int
award_points(char const *teamhash,
             char const *category,
             const long points,
             char const *uid)
{
  char   line[100];
  int    linelen;
  char   *filename;
  FILE   *f;
  time_t now = time(NULL);

  if (! team_exists(teamhash)) {
    return ERR_NOTEAM;
  }

  linelen = snprintf(line, sizeof(line),
                     "%s %s %ld %s",
                     teamhash, category, points, uid);
  if (sizeof(line) <= linelen) {
    return ERR_GENERAL;
  }

  if (anchored_search(state_path("points.log"), line, ' ')) {
    return ERR_CLAIMED;
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
     still able to record their score and move on.
  */

  filename = state_path("points.tmp/%lu.%d.%s.%s.%ld",
                        (unsigned long)now, getpid(),
                        teamhash, category, points);
  f = fopen(filename, "w");
  if (! f) {
    return ERR_GENERAL;
  }

  if (EOF == fprintf(f, "%lu %s\n", (unsigned long)now, line)) {
    return ERR_GENERAL;
  }

  fclose(f);

  /* Rename into points.new */
  {
    char ofn[PATH_MAX];

    strncpy(ofn, filename, sizeof(ofn));
    filename = state_path("points.new/%lu.%d.%s.%s.%ld",
                          (unsigned long)now, getpid(),
                          teamhash, category, points);
    rename(ofn, filename);
  }

  return 0;
}

