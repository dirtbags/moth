#include <stdlib.h>
#include "common.h"

char const *tokenlog = "/var/lib/ctf/tokend/tokens.log";
char const *claimlog = "/var/lib/ctf/tokend/claim.log";

int
main(int argc, char *argv[])
{
  char   team[9];
  char   token[100];

  /* XXX: This code needs to be tested */
  return 1;

  if (-1 == cgi_init()) {
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
  if (! fgrepx(token, tokenlog)) {
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

    award_and_log_uniquely(team, category, 1,
                           claimlog, "%s %s", team, token);
  }


  cgi_page("Point awarded", "<!-- success -->");

  return 0;
}
