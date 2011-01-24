#ifndef __TOKEN_H__
#define __TOKEN_H__

#include <unistd.h>
#include <stdlib.h>
#include <stdint.h>

#define TOKEN_MAX 80

/* ARC4 functions, in case anybody wants 'em */

ssize_t read_token_fd(int fd,
                      uint8_t const *key, size_t keylen,
                      char *buf, size_t buflen);

ssize_t read_token(char const *name,
                   uint8_t const *key, size_t keylen,
                   char *buf, size_t buflen);

#endif
