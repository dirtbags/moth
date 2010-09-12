#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>
#include "cgi.h"

static size_t inlen = 0;

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
cgi_page(char *title, char *fmt, ...)
{
  va_list  ap;

  printf(("Content-type: text/html\r\n"
          "\r\n"
          "<!DOCTYPE html>\n"
          "<html>\n"
          "  <head>\n"
          "    <title>%s</title>\n"
          "    <link rel=\"stylesheet\" href=\"ctf.css\" type=\"text/css\">\n"
          "  </head>\n"
          "  <body>\n"
          "    <h1>%s</h1>\n"),
         title, title);
  va_start(ap, fmt);
  vprintf(fmt, ap);
  va_end(ap);
  printf("\n"
         "  </body>\n"
         "</html>\n");
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
  exit(0);
}
