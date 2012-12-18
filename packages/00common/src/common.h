#ifndef __COMMON_H__
#define __COMMON_H__

#include <stddef.h>
#include <stdint.h>

#define TEAM_MAX 40
#define CAT_MAX 40
#define TOKEN_MAX 80
#define itokenlen 5

#define ERR_GENERAL -1
#define ERR_NOTEAM -2
#define ERR_CLAIMED -3

int cgi_init(char *global_argv[]);
size_t cgi_item(char *str, size_t maxlen);
void cgi_head(char *title);
void cgi_foot();
void cgi_result(int code, char *desc, char *fmt, ...);
void cgi_fail(int err);
void cgi_page(char *title, char *fmt, ...);
void cgi_error(char *text);


void ctf_chdir();
int anchored_search(char const *filename, char const *needle, const char anchor);
void urandom(char *buf, size_t buflen);
int my_snprintf(char *buf, size_t buflen, char *fmt, ...);
char *state_path(char const *fmt, ...);
char *package_path(char const *fmt, ...);
int team_exists(char const *teamhash);
int award_points(char const *teamhash,
                 char const *category,
                 long point,
                 char const *uid);

#endif
