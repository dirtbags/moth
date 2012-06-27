#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <stdio.h>
#include "cgi.h"

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
    printf("%d %s\r\n", code, desc);
  }
  cgi_head(desc);
  va_start(ap, fmt);
  vprintf(fmt, ap);
  va_end(ap);
  cgi_foot();
  exit(0);
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

