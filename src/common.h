#ifndef __COMMON_H__
#define __COMMON_H__

#define teamdir "/var/lib/ctf/teams/names"
#define pointsdir "/var/lib/ctf/points/new"

int fgrepx(char const *needle, char const *filename);
int team_exists(char const *teamhash);
int award_points(char const *teamhash,
                 char const *category,
                 int point);
void award_and_log_uniquely(char const *team,
                            char const *category,
                            int points,
                            char const *logfile,
                            char const *fmt, ...);

#endif
