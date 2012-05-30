#include <unistd.h>
#include <stdlib.h>

#define NOISE_PROB 300
#define NOISE_BITS 16

int badbits = 0;

char
line_noise(char c)
{
    int i = 7;

    while (badbits && (i >= 0)) {
        c = c ^ ((rand() % 2) << i);
        badbits -= 1;
        i -= 1;
    }

    if (rand() % NOISE_PROB == 0) {
        badbits = rand() % NOISE_BITS;
    }

    return c;
}

int
main(int argc, char *argv[])
{
    char c;
    ssize_t ret;
    int baud = 0;
    useconds_t usec;

    if (argv[1]) {
        baud = atoi(argv[1]);
    }
    if (! baud) {
        baud = 1200;
    }

    srandom(getpid());

    /* 
      N81 uses 1 stop bit, and 1 parity bit.  That works out to
       exactly 10 bits per byte.
     */
    usec = 10000000 / baud;

    while (1) {
        ret = read(0, &c, 1);
        if (ret != 1) {
            break;
        }
        c = line_noise(c);
        write(1, &c, 1);
        usleep(usec);
    }
    return 0;
}
