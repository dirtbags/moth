#ifndef __CGI_H__
#define __CGI_H__

#include <stddef.h>

int cgi_init();
size_t read_item(char *str, size_t maxlen);

#endif
