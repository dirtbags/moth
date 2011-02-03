#include <sys/types.h>
#include <sys/stat.h>
#include <errno.h>
#include <time.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <stddef.h>
#include <string.h>
#include <ctype.h>
#include <sysexits.h>
#include <dirent.h>
#include "common.h"
#include "arc4.h"
#include "md5.h"

void
list_teams()
{
  struct dirent *ent;
  DIR           *dir;

  dir = opendir(state_path("teams/names"));
  if (! dir) return;
  while ((ent = readdir(dir))) {
    struct stat buf;

    if ((0 == stat(state_path("teams/names/%s", ent->d_name), &buf)) &&
        (S_ISREG(buf.st_mode))) {
      char digest[MD5_HEXDIGEST_LEN + 1];

      md5_hexdigest((uint8_t *)ent->d_name, strlen(ent->d_name), digest);
      digest[MD5_HEXDIGEST_LEN] = '\n';
      write(1, digest, sizeof(digest));
    }
  }
  closedir(dir);
}

int
main(int argc, char *argv[])
{
  char    category[CAT_MAX];
  size_t  categorylen;
  char    token[TOKEN_MAX];
  size_t  tokenlen;
  uint8_t key[256];
  size_t  keylen;

  /* Read category name. */
  {
    ssize_t len;

    len = read(0, category, sizeof(category));
    if (0 >= len) return 0;

    /* Category name of "?" lists arc4 hashes of all teams */
    if ((1 == len) && ('?' == category[0])) {
      list_teams();
      return 0;
    }

    /* Strip invalid characters */
    for (categorylen = 0;
         (categorylen < len) && isalnum(category[categorylen]);
         categorylen += 1);
  }


  /* Read in that category's key. */
  {
    int fd;
    int ret;

    fd = open(package_path("mcp/tokend.keys/%.*s", (int)categorylen, category), O_RDONLY);
    if (-1 == fd) {
      fprintf(stderr, "Open key %.*s: %s\n",
              (int)categorylen, category, strerror(errno));
      return 0;
    }

    ret = read(fd, &key, sizeof(key));
    if (-1 == ret) {
      fprintf(stderr, "Read key %.*s: %s\n",
              (int)categorylen, category, strerror(errno));
      return 0;
    }
    keylen = (size_t)ret;

    close(fd);
  }

  /* Send a nonce, expect it back encrypted */
  {
    int32_t nonce;
    int32_t enonce = 0;

    urandom((char *)&nonce, sizeof(nonce));
    write(1, &nonce, sizeof(nonce));
    arc4_crypt_buffer(key, keylen, (uint8_t *)&nonce, sizeof(nonce));
    read(0, &enonce, sizeof(enonce));
    if (nonce != enonce) {
      write(1, ":<", 2);
      return 0;
    }
  }

  /* Create the token. */
  {
    unsigned char crap[itokenlen];
    unsigned char digest[bubblebabble_len(itokenlen)];

    urandom((char *)crap, sizeof(crap));

    /* Digest some random junk. */
    bubblebabble(digest, (unsigned char *)&crap, itokenlen);

    /* Append digest to category name. */
    tokenlen = (size_t)snprintf(token, sizeof(token),
                               "%.*s:%s",
                                (int)categorylen, category, digest);
  }

  /* Write that token out now. */
  {
    int fd;
    int ret;

    do {
      fd = open(state_path("tokens.db"), O_WRONLY | O_CREAT, 0666);
      if (-1 == fd) break;

      ret = lockf(fd, F_LOCK, 0);
      if (-1 == ret) break;

      ret = lseek(fd, 0, SEEK_END);
      if (-1 == ret) break;

      ret = write(fd, token, tokenlen);
      if (-1 == ret) break;

      ret = write(fd, "\n", 1);
      if (-1 == ret) break;

      ret = close(fd);
      if (-1 == ret) break;
    } while (0);

    if ((-1 == fd) || (-1 == ret)) {
      printf("!%s", strerror(errno));
      return 0;
    }
  }

  /* Encrypt the token. */
  {
    arc4_crypt_buffer(key, keylen, (uint8_t *)token, tokenlen);
  }

  /* Send it back.  If there's an error here, it's okay.  Better to have
     unclaimed tokens than unclaimable ones. */
  write(1, token, tokenlen);

  return 0;
}
