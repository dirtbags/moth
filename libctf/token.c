#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stddef.h>
#include <stdlib.h>
#include <unistd.h>
#include <values.h>

#ifndef CTF_BASE
#define CTF_BASE "/var/lib/ctf"
#endif

ssize_t
read_token_fd(int fd,
              uint8_t const *key, size_t keylen,
              char *buf, size_t buflen)
{
  ssize_t ret;

  ret = read(fd, buf, buflen);
  if (-1 != ret) {
    arc4_crypt_buffer(key, keylen, (uint8_t *)buf, (size_t)ret);
  }
  return ret;
}


ssize_t
read_token(char const *name,
           uint8_t const *key, size_t keylen,
           char *buf, size_t buflen)
{
  char    path[PATH_MAX];
  int     pathlen;
  int     fd;
  ssize_t ret;

  pathlen = snprintf(path, sizeof(path) - 1,
                     CTF_BASE "/tokens/%s", name);
  path[pathlen] = '\0';

  fd = open(path, O_RDONLY);
  if (-1 == fd) return -1;
  ret = read_token_fd(fd, key, keylen, buf, buflen);
  close(fd);
  return ret;
}
