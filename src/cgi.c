#include <stdlib.h>
#include <stdio.h>
#include "cgi.h"

static size_t inlen = 0;

int
cgi_init()
{
  char *rm = getenv("REQUEST_METHOD");

  if (! (rm && (0 == strcmp(rm, "POST")))) {
    printf("405 Method not allowed\n"
           "Allow: POST\n"
           "Content-type: text/html\n"
           "\n"
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
read_item(char *str, size_t maxlen)
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
