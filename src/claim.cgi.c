#include <stdlib.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  char   team[TEAM_MAX]   = {0};
  char   token[TOKEN_MAX] = {0};

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
    cgi_page("No such team", "");
  }

  /* Any weird characters in token name? */
  {
    char *p;

    for (p = token; *p; p += 1) {
      if ((! isalnum(*p)) &&
          (*p != '-') &&
          (*p != ':')) {
        cgi_page("Invalid token", "");
      }
    }
  }


  /* Does the token exist? */
  if (! fgrepx(token, srv_path("tokens.db"))) {
    cgi_page("Token does not exist", "");
  }

  /* Award points */
  {
    char category[40];
    int  i;

    /* Pull category name out of the token */
    for (i = 0; token[i] != ':'; i += 1) {
      category[i] = token[i];
    }
    category[i] = '\0';

    {
      char line[TEAM_MAX + TOKEN_MAX + 1];

      my_snprintf(line, sizeof(line),
                  "%s %s", team, token);
      award_and_log_uniquely(team, category, 1,
                             "tokens.db", line);
    }
  }


  cgi_page("Point awarded", "<!-- success -->");

  return 0;
}
