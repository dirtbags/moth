#ifndef __COMMON_H__
#define __COMMON_H__

#include <stddef.h>

#define TEAM_MAX 40
#define CAT_MAX 40

int cgi_init(char *global_argv[]);
size_t cgi_item(char *str, size_t maxlen);
void cgi_head(char *title);
void cgi_foot();
void cgi_page(char *title, char *fmt, ...);
void cgi_error(char *fmt, ...);


int fgrepx(char const *needle, char const *filename);
char *srv_path(char const *fmt, ...);
int team_exists(char const *teamhash);
int award_points(char const *teamhacsh,
                 char const *category,
                 long point);
void award_and_log_uniquely(char const *team,
                            char const *category,
                            long points,
                            char const *logfile,
                            char const *fmt, ...);

#endif
