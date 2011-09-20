#include <stdlib.h>
#include <ctype.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  char team[TEAM_MAX]   = {0};
  char token[TOKEN_MAX] = {0};

  if (-1 == cgi_init(argv)) {
    return 0;
  }

  /* Read in team and token */
  while (1) {
    size_t len;
    char   key[20];

    len = cgi_item(key, sizeof(key));
    if (0 == len) break;
    switch (key[0]) {
      case 't':
        cgi_item(team, sizeof(team));
        break;
      case 'k':
        cgi_item(token, sizeof(token));
        break;
    }
  }

  if (! team_exists(team)) {
    cgi_result(409, "No such team", "<p>There is no team with that hash.</p>");
  }

  /* Any weird characters in token name? */
  {
    char *p;

    if ('\0' == token[0]) {
      cgi_result(409, "Must supply token", "<p>Your request did not contain a k= parameter.</p>");
    }
    for (p = token; *p; p += 1) {
      if ((! isalnum(*p)) &&
          (*p != '-') &&
          (*p != ':')) {
        cgi_result(409, "Not a token", "<p>This token has untokenlike characteristics.</p>");
      }
    }
  }


  /* Does the token exist? */
  if (! fgrepx(token, state_path("tokens.db"))) {
    cgi_result(409, "No such token", "<p>This token has not been issued.</p>");
  }

  /* Award points */
  {
    char *p = token;
    char *q;
    char  category[40];
    char  points_s[40];
    int   points;

    /* Pull category name out of the token */
    for (q = category; *p && (*p != ':'); p += 1) {
      *(q++) = *p;
    }
    *q = '\0';
    if (p) p += 1;

    /* Pull point value out of the token (if it has one) */
    for (q = points_s; *p && (*p != ':'); p += 1) {
      *(q++) = *p;
    }
    *q = '\0';
    points = atoi(points_s);
    if (0 == points) points = 1;

    {
      char line[200];

      my_snprintf(line, sizeof(line), "%s %s", team, token);
      award_and_log_uniquely(team, category, points, state_path("claim.db"), line);
    }
  }

  cgi_page("Point awarded", "<p>Congratulations.</p>");

  return 0;
}
