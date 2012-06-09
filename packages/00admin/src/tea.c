#include <stdio.h>
#include <unistd.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <sysexits.h>

static const char magic[4] = "tea.";

#define min(a,b) (((a)<(b))?(a):(b))

void 
tea_encrypt(uint32_t *v, uint32_t *k) {
    uint32_t v0=v[0], v1=v[1], sum=0, i;           /* set up */
    uint32_t delta=0x9e3779b9;                     /* a key schedule constant */
    uint32_t k0=k[0], k1=k[1], k2=k[2], k3=k[3];   /* cache key */
    for (i=0; i < 32; i++) {                       /* basic cycle start */
        sum += delta;
        v0 += ((v1<<4) + k0) ^ (v1 + sum) ^ ((v1>>5) + k1);
        v1 += ((v0<<4) + k2) ^ (v0 + sum) ^ ((v0>>5) + k3);  
    }                                              /* end cycle */
    v[0]=v0; v[1]=v1;
}

/*
 * Use TEA in CTR mode to create a stream cipher.
 */
static void
tea_apply(FILE *out, FILE *in, uint32_t *k, uint32_t ivec)
{
    uint32_t count = 0;
    uint32_t v[2];
    size_t idx = sizeof v;
    char *p    = (char *)v;

    while (1) {
        int c = fgetc(in);
        
        if (EOF == c) {
            break;
        }

        if (sizeof v == idx) {
            v[0] = ivec;
            v[1] = count++;
            tea_encrypt(v, k);
            idx = 0;
        }

        fputc(c ^ p[idx++], out);
    }
}

int
tea_decrypt_stream(FILE *out, FILE *in, uint32_t *k)
{
    uint32_t ivec;

   {
        char m[4] = {0};

        fread(m, sizeof m, 1, in);
        if (memcmp(m, magic, 4)) {
            return -1;
        }
    }

    fread(&ivec, sizeof ivec, 1, in);
    tea_apply(out, in, k, ivec);

    return 0;
}

int
tea_encrypt_stream(FILE *out, FILE *in, uint32_t *k)
{
    uint32_t ivec;

    fwrite(magic, sizeof magic, 1, out);

    {
        FILE *r = fopen("/dev/urandom", "r");

        if (! r) {
            return -1;
        }
        fread(&ivec, sizeof ivec, 1, r);
        fclose(r);
    }
    fwrite(&ivec, sizeof ivec, 1, out);

    tea_apply(out, in, k, ivec);

    return 0;
}

int
usage(const char *prog)
{
  fprintf(stderr, "Usage: %s [-e] <PLAINTEXT\n", prog);
  fprintf(stderr, "\n");
  fprintf(stderr, "You must pass in a key on fd 3 or in the environment variable KEY.\n");
  return EX_USAGE;
}

int
main(int argc, char *argv[])
{
    uint32_t key[4] = {0};

    {
        char *ekey = getenv("KEY");

        if (ekey) {
            memcpy(key, ekey, min(strlen(ekey), sizeof(key)));
        } else if (-1 == read(3, key, sizeof(key))) {
            return usage(argv[0]);
        }
    }

    if (! argv[1]) {
        tea_decrypt_stream(stdout, stdin, key);
    } else if (0 == strcmp(argv[1], "-e")) {
        tea_encrypt_stream(stdout, stdin, key);
    } else {
        return usage(argv[0]);
    }

  return 0;
}
