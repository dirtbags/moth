#include <stdio.h>
#include <sys/time.h>

void
mungle(char *str, int len)
{
    int i;

    for (i = 0; i < len; i += 1) {
        str[i] ^= 0xff;
    }
}

int
main(int argc, char *argv[])
{
    long answer = 0;
    int i;

    {
        struct timeval tv;

        gettimeofday(&tv, NULL);
        srandom(tv.tv_usec);
    }

    for (i = 0; i < 4; i += 1) {
        answer = (answer << 4) | ((random() % 6) + 1);
    }

    while (1) {
        char line[20];
        long guess;
        int ret = 0;

        if (NULL == fgets(line, sizeof(line), stdin)) {
            break;
        }

        guess = strtol(line, NULL, 16);

        for (i = 0; i < 4; i += 1) {
            int g = (guess  >> (i*4)) & 0xf;
            int a = (answer >> (i*4)) & 0xf;

            if ((g < 1) || (g > 7)) {
                ret = 0;
                break;
            } else if (g == a) {
                ret += 0x10;
            } else if (g & a) {
                ret += 0x01;
            }
        }

        printf("%02x\n", ret);
    }

    return 0;
}
