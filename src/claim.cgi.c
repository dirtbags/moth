#include <stdio.h>
#include <unistd.h>
#include <time.h>
#include "cgi.h"
#include "pointscli.h"

char const *tokenlog = "/var/lib/ctf/tokend/tokens.log";
char const *teamdir = "/var/lib/ctf/teams";
char const *claimlog = "/var/lib/ctf/tokend/claim.log";

int
mystrcmp(char *a, char *b)
{
  while (*a == *b) {
    a += 1;
    b += 1;
  }
  if (((*a == '\0') || (*a == '\n')) &&
      ((*b == '\0') || (*b == '\n'))) {
    return 0;
  }
  return -1;
}

int
main(int argc, char *argv[])
{
  char   team[9];
  size_t teamlen;
  char   token[100];
  size_t tokenlen;

  if (-1 == cgi_init()) {
    return 0;
  }

  /* Read in team and token */
  while (1) {
    size_t len;
    char   key[20];
    int    i;

    len = cgi_item(key, sizeof(key));
    if (0 == len) break;
    if (1 == len) {
      if ('t' == key[0]) {
        teamlen = cgi_item(team, sizeof(team) - 1);

        /* End string at # read or first bad char */
        for (i = 0; i < teamlen; i += 1) {
          if (! isalnum(team[i])) {
            break;
          }
        }
        team[i] = '\0';
      } else if ('k' == key[0]) {
        tokenlen = cgi_item(token, sizeof(token) - 1);

        /* End string at # read or first bad char */
        for (i = 0; i < tokenlen; i += 1) {
          if ((! isalnum(token[i])) &&
              (token[i] != '-') &&
              (token[i] != ':')) {
            break;
          }
        }
        token[i] = '\0';
      }
    }
  }

  if (! pointscli_isteam(team)) {
    cgi_page("No such team", "");
  }

  /* Does the token exist? */
  {
    FILE *f;
    int   valid = 0;

    f = fopen(filelog, "r");
    if (f) {
      while (1) {
        char line[100];

        if (NULL == fgets(line, sizeof(line), f)) {
          break;
        }
        if (0 == mystrcmp(line, token)) {
          valid = 1;
          break;
        }
      }
      fclose(f);
    }

    if (! valid) {
      cgi_page("Invalid token",
               "<p>Sorry, that token's no good.</p>");
    }
  }

  /* If the token's unclaimed, award points and log the claim */
  {
    FILE *f;
    int   claimed = 0;
    char  needle[100];

    f = fopen(claimlog, "rw");
    if (! f) {
      cgi_error("Couldn't fopen(\"%s\", \"rw\")", claimlog);
    }
    sprintf(needle, "%s %s", team, token);
    while (1) {
      char line[100];
      int  pos;

      if (NULL == fgets(line, sizeof(line), f)) {
        break;
      }

      /* Skip to past first space */
      for (; (*line && (*line != ' ')); line += 1);

      if (0 == mystrcmp(line, needle)) {
        claimed = 1;
        break;
      }
    }
    if (claimed) {
      cgi_page("Already claimed",
               "<p>Apparently you've already claimed that token.</p>");
    }

    /* Now register the points */
    {
      char category[20];
      int  i;

      /* Pull category name out of the token */
      for (i = 0; token[i] != ':'; i += 1) {
        category[i] = token[i];
      }
      category[i] = '\0';

      pointscli_award(team, category, points);
    }

    /* Finally, append an entry to the log file.  I figure it's better
       to give points first and fail here, than it is to lock them out
       of making points and then fail to award them. */
    {
      char       timestamp[40];
      time_t     t;
      struct tm *tmp;
      int        ret;

      t = time(NULL);
      tmp = localtime(&t);
      if (NULL == tmp) {
        cgi_error("I... uh... couldn't figure out what time it is.");
      }
      if (0 == strftime(timestamp, sizeof(timestamp), "%Y-%m-%dT%H:%M:%S", tmp)) {
        cgi_error("I forgot how to format time.");
      }
      fseek(f, 0, SEEK_SET);
      if (-1 == lockf(fileno(f), F_LOCK, 0)) {
        cgi_error("Unable to lock the log file.");
      }
      fseek(f, 0, SEEK_END);
      fprintf(f, "%s %s %s\n", timestamp, team, token);
    }

    fclose(f);
  }

  cgi_page("Points awarded", "<!-- success -->");

  return 0;
}
