#ifndef __TOKEN_H__
#define __TOKEN_H__

#include <unistd.h>
#include <stdlib.h>
#include <stdint.h>

#define TOKEN_MAX 80

/* ARC4 functions, in case anybody wants 'em */
struct arc4_ctx;
void arc4_init(struct arc4_ctx *ctx,
               uint8_t const *key, size_t keylen);
void arc4_crypt(struct arc4_ctx *ctx,
                uint8_t *obuf, uint8_t const *ibuf, size_t buflen);
void arc4_crypt_buffer(uint8_t const *key, size_t keylen,
                       uint8_t *buf, size_t buflen);

ssize_t read_token_fd(int fd,
                      uint8_t const *key, size_t keylen,
                      char *buf, size_t buflen);

ssize_t read_token(char const *name,
                   uint8_t const *key, size_t keylen,
                   char *buf, size_t buflen);

#endif
