#include <stdio.h>
#include <stdlib.h>
#include <sysexits.h>
#include <time.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  int points;
  int ret;
  char comment[512];

  if (argc != 5) {
    fprintf(stderr, "Usage: pointscli TEAM CATEGORY POINTS 'COMMENT'\n");
    return EX_USAGE;
  }

  points = atoi(argv[3]);
  if (0 == points) {
    fprintf(stderr, "Error: award 0 points?\n");
    return EX_USAGE;
  }

  snprintf(comment, sizeof comment, "--%s", argv[4]);

  ret = award_points(argv[1], argv[2], points, comment);
  switch (ret) {
    case ERR_GENERAL:
      perror("General error");
      return EX_UNAVAILABLE;
    case ERR_NOTEAM:
      fprintf(stderr, "No such team\n");
      return EX_NOUSER;
    case ERR_CLAIMED:
      fprintf(stderr, "Duplicate entry\n");
      return EX_DATAERR;
    default:
      fprintf(stderr, "Error %d\n", ret);
      return EX_SOFTWARE;
  }

  return 0;
}
