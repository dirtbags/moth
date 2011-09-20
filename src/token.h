#ifndef __TOKEN_H__
#define __TOKEN_H__

#define TOKEN_MAX 50

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

ssize_t write_token(FILE *out,
                    const char *name,
                    const uint8_t *key, size_t keylen);
ssize_t print_token(const char *name,
                    const uint8_t *key, size_t keylen);
ssize_t get_token(char *buf, size_t buflen,
                  const char *name,
                  const uint8_t *key, size_t keylen);

#endif
