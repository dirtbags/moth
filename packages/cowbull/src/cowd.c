#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <time.h>
#include <string.h>
#include <sysexits.h>
#include <arpa/inet.h>
#include <sys/socket.h>

#define TIMEOUT 30

#define NTOKENS 20
#define TOKENLEN 50
char            tokens[NTOKENS][TOKENLEN];
int             ntokens;

struct state {
    time_t          death;
    uint16_t        answer;
    uint16_t        guesses;
};

#define NSTATES 500
struct state    states[NSTATES] = { 0 };

int
bind_port(struct in6_addr *addr, int fd, uint16_t port)
{
    struct sockaddr_in6 saddr = { 0 };

    saddr.sin6_family = AF_INET6;
    saddr.sin6_port = htons(port);
    memcpy(&saddr.sin6_addr, addr, sizeof *addr);
    return bind(fd, (struct sockaddr *) &saddr, sizeof saddr);
}


struct newgame {
    uint16_t        offset;
    uint16_t        token;
};

void
new_game(int sock, time_t now, struct sockaddr_in6 *from,
         socklen_t fromlen)
{
    int             i;
    struct newgame  g;

    for (g.offset = 0; g.offset < NSTATES; g.offset += 1) {
        struct state   *s = &states[g.offset];

        if (s->death < now) {
            s->death = now + TIMEOUT;
            s->guesses = 0;
            s->answer = 0;

            for (i = 0; i < 4; i += 1) {
                s->answer = (s->answer << 4) | ((random() % 6) + 1);
            }
            break;
        }
    }

    if (g.offset < NSTATES) {
        sendto(sock, &g, sizeof(g), 0, (struct sockaddr *) from, fromlen);
    }
}

struct guess {
    uint16_t        offset;
    uint16_t        token;
    uint16_t        guess;
};

void
loop(int sock)
{
    struct guess    g;
    struct state   *cur;
    struct sockaddr_in6 from;
    socklen_t       fromlen = sizeof from;
    time_t          now = time(NULL);

    /*
     * Read guess 
     */
    {
        ssize_t         inlen;

        inlen = recvfrom(sock, &g, sizeof g, 0,
                         (struct sockaddr *) &from, &fromlen);
        if (inlen != sizeof g) {
            return;
        }
    }

    /*
     * Bounds check 
     */
    if (g.offset >= NSTATES) {
        g.offset = 0;
    }
    cur = &states[g.offset];

    if ((g.token != cur->answer) || /* Wrong token? */
        (cur->death < now) ||   /* Old game? */
        (cur->guesses++ > 100)) {       /* Too dumb? */
        /*
         * Start a new game 
         */
        new_game(sock, now, &from, fromlen);
        return;
    } else {
        uint8_t         reply;
        int i;

        for (i = 0; i < 4; i += 1) {
            int             s = (g.guess >> (i * 4)) & 0xf;
            int             a = (cur->answer >> (i * 4)) & 0xf;
            if ((s < 1) || (s > 7)) {
                reply = 0;
                break;
            } else if (s == a) {
                reply += 0x10;
            } else if (s & a) {
                reply += 0x01;
            }
        }

        if (reply == 0x40) {
            if (cur->guesses > ntokens) {
                sendto(sock, tokens[cur->guesses],
                       strlen(tokens[cur->guesses]), 0,
                       (struct sockaddr *) &from, fromlen);
            }
        } else {
            sendto(sock, &reply, sizeof reply, 0, (struct sockaddr *) &from,
                   fromlen);
        }
    }
}

int
main(int argc, char *argv[])
{
    long            answer = 0;
    int             sock;
    int             i;
    struct in6_addr addr;

    srand(time(NULL));

    if (argc > 1) {
        if (0 >= inet_pton(AF_INET6, argv[1], &addr)) {
            fprintf(stderr, "invalid address: %s\n", argv[1]);
            return EX_IOERR;
        }
    } else {
        memcpy(&addr, &in6addr_any, sizeof addr);
    }

    /*
     * Read in tokens 
     */
    for (ntokens = 0; ntokens < NTOKENS; ntokens += 1) {
        if (NULL == fgets(tokens[ntokens], TOKENLEN, stdin)) {
            break;
        }
    }
    printf("Read %d tokens.\n", ntokens);

    /*
     * Set up socket 
     */
    sock = socket(AF_INET6, SOCK_DGRAM, 0);
    i = bind_port(&addr, sock, 3782);
    if (-1 == i) {
        perror("Bind port 3782");
        return EX_IOERR;
    }

    for (i = 0; i < 4; i += 1) {
        answer = (answer << 4) | ((random() % 6) + 1);
    }

    while (1) {
        loop(sock);
    }

    return 0;
}
