#ifndef __COMMON_H__
#define __COMMON_H__

#define teamdir "/var/lib/ctf/teams"
#define pointslog "/var/lib/ctf/points.log"

int team_exists(char *teamhash);
int award_points(char *teamhash, char *category, int point);

#endif
