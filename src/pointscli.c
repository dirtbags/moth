#include <stdio.h>
#include <stdlib.h>
#include <sysexits.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  int points;
  int ret;

  if (argc != 4) {
    fprintf(stderr, "Usage: pointscli TEAM CATEGORY POINTS\n");
    return EX_USAGE;
  }

  points = atoi(argv[3]);
  if (0 == points) {
    fprintf(stderr, "Error: award 0 points?\n");
    return EX_USAGE;
  }

  ret = award_points(argv[1], argv[2], points);
  switch (ret) {
    case -3:
      fprintf(stderr, "Runtime error\n");
      return EX_SOFTWARE;
    case -2:
      fprintf(stderr, "No such team\n");
      return EX_NOUSER;
    case -1:
      perror("Couldn't award points");
      return EX_UNAVAILABLE;
  }

  return 0;
}
