#include <stdlib.h>
#include "common.h"

char const *logfile = "/var/lib/ctf/puzzler.log";

int
main(int argc, char *argv)
{
  char team[TEAM_MAX];
  char category[CAT_MAX];
  char points_str[5];
  char answer[500];
  long points;

  if (-1 == cgi_init()) {
    return 0;
  }

  /* Read in team and answer */
  while (1) {
    size_t len;
    char   key[20];

    len = cgi_item(key, sizeof(key));
    if (0 == len) break;
    switch (key[0]) {
      case 't':
        cgi_item(team, sizeof(team));
        break;
      case 'c':
        cgi_item(category, sizeof(category));
        break;
      case 'p':
        cgi_item(points_str, sizeof(points_str));
        points = atol(points_str);
        break;
      case 'a':
        cgi_item(answer, sizeof(answer));
        break;
    }
  }

  /* Check to see if team exists */
  if (! team_exists(team)) {
    cgi_page("No such team", "");
  }

  /* Validate category name (prevent directory traversal) */
  {
    char *p;

    for (p = category; *p; p += 1) {
      if (! isalnum(*p)) {
        cgi_page("Invalid category", "");
      }
    }
  }

  /* Check answer (also assures category exists) */
  {
    char filename[100];
    char needle[400];

    my_snprintf(filename, sizeof(filename),
                "/srv/%s/answers.txt", category);
    my_snprintf(needle, sizeof(needle),
                "%ld %s", points, answer);
    if (! fgrepx(needle, filename)) {
      cgi_page("Wrong answer", "");
    }
  }

  award_and_log_uniquely(team, category, points,
                         logfile, "%s %s %ld", team, category, points);

  cgi_page("Points awarded",
           ("<p>%d points for %s.</p>\n"
            "<p>Patience please: it may be up to 1 minute before "
            "the next puzzle opens in this category.</p>"),
           points, team);

  return 0;
}
