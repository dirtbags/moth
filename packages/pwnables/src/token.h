#ifndef __TOKEN_H__
#define __TOKEN_H__

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

ssize_t write_token(FILE *out,
                    const char *name,
                    const uint8_t *key, size_t keylen);
ssize_t print_token(const char *name,
                    const uint8_t *key, size_t keylen);


#endif
