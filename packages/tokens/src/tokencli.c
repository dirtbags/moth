#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <string.h>
#include <unistd.h>
#include <sysexits.h>
#include <stdio.h>
#include "arc4.h"

/* I don't feel compelled to put all the TCP client code in here
 * when it's so simple to run this with netcat or ucspi.  Plus, using
 * stdin and stdout makes it simpler to test.
 */

int
read_key(char *filename, uint8_t *key, size_t *keylen)
{
  int fd = open(filename, O_RDONLY);
  int len;

  if (-1 == fd) {
    perror("open");
    return EX_NOINPUT;
  }

  len = read(fd, key, *keylen);
  if (-1 == len) {
    perror("read");
    return EX_NOINPUT;
  }
  *keylen = (size_t)len;

  return 0;
}

int
main(int argc, char *argv[]) {
  uint8_t skey[200];
  size_t  skeylen = sizeof(skey);
  char    token[200];
  size_t  tokenlen;
  int     ret;

  if (argc != 3) {
    fprintf(stderr, "Usage: %s SERVICE SERVICEKEY 3>TOKENFILE\n", argv[0]);
    fprintf(stderr, "\n");
    fprintf(stderr, "SERVICEKEY is a filename.\n");
    fprintf(stderr, "Server chatter happens over stdin and stdout.\n");
    fprintf(stderr, "Tokens are written to file descriptor 3.\n");
    fprintf(stderr, "\n");
    fprintf(stderr, "To run with netcat:\n");
    fprintf(stderr, "    nc server 1 -e tokencli cat cat.key 3> tokenfile\n");
    return EX_USAGE;
  }

  /* read in keys */
  ret = read_key(argv[2], skey, &skeylen);
  if (0 != ret) return ret;

  /* write service name */
  write(1, argv[1], strlen(argv[1]));

  /* read nonce, send back encrypted version */
  {
    uint8_t nonce[80];
    int     noncelen;

    noncelen = read(0, nonce, sizeof(nonce));
    if (0 >= noncelen) {
      perror("read");
      return EX_IOERR;
    }
    arc4_crypt_buffer(skey, skeylen, nonce, (size_t)noncelen);
    write(1, nonce, (size_t)noncelen);
  }

  /* read token */
  {
    int len;

    len = read(0, token, sizeof(token));
    if (0 >= len) {
      perror("read");
      return EX_IOERR;
    }
    tokenlen = (size_t)len;
  }

  /* decrypt it */
  arc4_crypt_buffer(skey, skeylen, (uint8_t *)token, tokenlen);

  /* write it to fd 3 */
  write(3, token, tokenlen);

  return 0;
}
