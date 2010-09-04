#ifndef __COMMON_H__
#define __COMMON_H__

#define teamdir "/var/lib/ctf/teams"
#define pointslog "/var/lib/ctf/points.log"

int timestamp(char *now, size_t nowlen);
int team_exists(char *teamhash);
int award_points(char *teamhash, char *category, int point);

#endif
