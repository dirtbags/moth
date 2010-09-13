#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <dirent.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "common.h"

int
longcmp(long *a, long *b)
{
  if (*a < *b) return -1;
  if (*a > *b) return 1;
  return 0;
}

#define PUZZLES_MAX 500

/** Keeps track of the most points yet awarded in each category */
struct {
  char cat[CAT_MAX];
  long points;
} points_by_cat[100];
int ncats = 0;

void
read_points_by_cat()
{
  FILE *f = fopen(srv_path("puzzler.db"), "r");
  char  cat[CAT_MAX];
  long  points;
  int   i;

  if (! f) {
    return;
  }

  while (1) {
    if (2 != fscanf(f, "%*s %s %ld\n", &cat, &points)) {
      break;
    }
    for (i = 0; i < ncats; i += 1) {
      if (0 == strcmp(cat, points_by_cat[i].cat)) break;
    }
    if (i == ncats) {
      strcpy(points_by_cat[i].cat, cat);
      ncats += 1;
    }
    if (points > points_by_cat[i].points) {
      points_by_cat[i].points = points;
    }
  }
}

int
main(int argc, char *argv[])
{
  int i;
  DIR *srv;

  if (-1 == cgi_init(argv)) {
    return 0;
  }

  read_points_by_cat();

  srv = opendir(srv_path("packages"));
  if (NULL == srv) {
    cgi_error("Cannot opendir(\"/srv\")");
  }

  cgi_head("Open puzzles");
  printf("<dl>\n");

  /* For each file in /srv/ ... */
  while (1) {
    struct dirent *e          = readdir(srv);
    char          *cat        = e->d_name;
    DIR           *puzzles;
    long           catpoints[PUZZLES_MAX];
    size_t         ncatpoints = 0;

    if (! e) break;
    if ('.' == cat[0]) continue;
    /* We have to lstat anyway to see if it's a directory; may as
       well just barge ahead and watch for errors. */

    /* Open /srv/ctf/$cat/puzzles/ */
    puzzles = opendir(srv_path("packages/%s/puzzles", cat));
    if (NULL == puzzles) {
      continue;
    }

    while (ncatpoints < PUZZLES_MAX) {
      struct dirent *pe = readdir(puzzles);
      long           points;
      char          *p;

      if (! pe) break;

      /* Only do this if it's an int */
      points = strtol(pe->d_name, &p, 10);
      if (*p) continue;

      catpoints[ncatpoints++] = points;
    }

    closedir(puzzles);

    /* Sort points */
    qsort(catpoints, ncatpoints, sizeof(*catpoints),
          (int (*)(const void *, const void *))longcmp);


    /* Print out point values up to one past the last solved puzzle in
       this category */
    {
      long maxpoints = 0;

      /* Find the most points scored in this category */
      for (i = 0; i < ncats; i += 1) {
        if (0 == strcmp(cat, points_by_cat[i].cat)) {
          maxpoints = points_by_cat[i].points;
          break;
        }
      }

      printf("  <dt>%s</dt>\n", cat);
      printf("  <dd>\n");
      for (i = 0; i < ncatpoints; i += 1) {
        printf("    <a href=\"/puzzles/%s/%d\">%d</a>\n",
               cat, catpoints[i], catpoints[i]);
        if (catpoints[i] > maxpoints) break;
      }
      printf("  </dd>\n");
    }
  }

  closedir(srv);

  printf("</dl>\n");
  cgi_foot();

  return 0;
}
