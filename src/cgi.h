#ifndef __CGI_H__
#define __CGI_H__

#include <stddef.h>

int cgi_init();
size_t cgi_item(char *str, size_t maxlen);
void cgi_page(char *title, char *fmt, ...);
void cgi_error(char *fmt, ...);

#endif
