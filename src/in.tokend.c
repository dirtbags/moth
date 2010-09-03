#include <sys/types.h>
#include <time.h>
#include <unistd.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <stddef.h>
#include <ctype.h>
#include "xxtea.h"

#define itokenlen 3

char const *keydir = "/var/lib/ctf/tokend/keys";
char const *tokendir = "/var/lib/ctf/tokend/tokens";

char const consonants[] = "bcdfghklmnprstvz";
char const vowels[]     = "aeiouy";

#define bubblebabble_len(n) (6*(((n)/2)+1))

/** Compute bubble babble for input buffer.
 *
 * The generated output will be of length 6*((inlen/2)+1), including the
 * trailing NULL.
 *
 * Test vectors:
 *     `' (empty string) `xexax'
 *     `1234567890'      `xesef-disof-gytuf-katof-movif-baxux'
 *     `Pineapple'       `xigak-nyryk-humil-bosek-sonax'
 */
void
bubblebabble(char *out, char const *in, const size_t inlen)
{
  size_t pos  = 0;
  int    seed = 1;
  size_t i    = 0;

  out[pos++] = 'x';
  while (1) {
    unsigned char c;

    if (i == inlen) {
      out[pos++] = vowels[seed % 6];
      out[pos++] = 'x';
      out[pos++] = vowels[seed / 6];
      break;
    }

    c = in[i++];
    out[pos++] = vowels[(((c >> 6) & 3) + seed) % 6];
    out[pos++] = consonants[(c >> 2) & 15];
    out[pos++] = vowels[((c & 3) + (seed / 6)) % 6];
    if (i == inlen) {
      break;
    }
    seed = ((seed * 5) + (c * 7) + in[i]) % 36;

    c = in[i++];
    out[pos++] = consonants[(c >> 4) & 15];
    out[pos++] = '-';
    out[pos++] = consonants[c & 15];
  }

  out[pos++] = 'x';
  out[pos] = '\0';
}

int
main(int argc, char *argv[])
{
  char     service[50];
  char     token[80];
  uint32_t key[4];
  size_t   tokenlen;

  /* Seed the random number generator.  This ought to be unpredictable
     enough for a contest. */
  srand((int)time(NULL) * (int)getpid());

  /* Read service name. */
  {
    size_t len;
    int    i;

    len = fread(service, 1, sizeof(service) - 1, stdin);
    for (i = 0; (i < len) && isalnum(service[i]); i += 1);
    service[i] = '\0';
  }

  /* Read in that service's key. */
  {
    char    path[100];
    FILE   *f = NULL;
    size_t  len;
    int     ret;

    ret = snprintf(path, sizeof(path),
                   "%s/%s", keydir, service);
    if (ret < sizeof(path)) {
      f = fopen(path, "r");
    }
    if (! f) {
      printf("!Unregistered service");
      return 0;
    }

    len = fread(&key, sizeof(uint32_t), 4, f);
    fclose(f);

    if (4 != len) {
      printf("!Key file too short");
      return 0;
    }
  }

  /* Create the token. */
  {
    uint8_t crap[itokenlen];
    char    digest[bubblebabble_len(itokenlen)];
    int     i;

    /* Digest some random junk. */
    for (i = 0; i < itokenlen; i += 1) {
      crap[i] = (uint8_t)random();
    }
    bubblebabble(digest, crap, itokenlen);

    /* Append digest to service name.  I use . as a separator because it
       won't be URL encoded. */
    tokenlen = (size_t)snprintf(token, sizeof(token),
                                "%s.%s",
                                service, digest);
  }

  /* Write that token out now. */
  {
    char    path[100];
    FILE   *f = NULL;
    int     ret;

    ret = snprintf(path, sizeof(path),
                   "%s/%s", tokendir, token);
    f = fopen(path, "w");
    if (f) {
      fclose(f);
    } else {
      printf("!Unable to write token");
      return 0;
    }
  }

  /* Encrypt the token.  Note that now tokenlen is in uint32_ts, not
     chars! */
  {
    tokenlen = (tokenlen + (tokenlen % 4)) / 4;

    tea_encode(key, (uint32_t *)token, tokenlen);
  }

  /* Send it back.  If there's an error here, it's okay.  Better to have
     unclaimed tokens than unclaimable ones. */
  fwrite(token, tokenlen, sizeof(uint32_t), stdout);

  return 0;
}
