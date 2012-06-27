#ifndef __CGI_H_
#define __CGI_H_

#include <stddef.h>

int cgi_init(char *global_argv[]);
size_t cgi_item(char *str, size_t maxlen);
void cgi_head(char *title);
void cgi_foot();
void cgi_result(int code, char *desc, char *fmt, ...);
void cgi_page(char *title, char *fmt, ...);
void cgi_error(char *text);

#endif
