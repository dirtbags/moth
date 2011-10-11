/* multicast_server.c
 * Adapted from tmouse's IPv6 client/server example code
 * found at http://cboard.cprogramming.com/showthread.php?t=67469
 */

#include <stdio.h>      /* for fprintf() */
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <stdlib.h>     /* for atoi() and exit() */
#include <unistd.h>
#include <string.h>

static void DieWithError(const char* errorMessage)
{
    fprintf(stderr, "%s\n", errorMessage);
    exit(10);
}

int main(int argc, char *argv[])
{
    int    sock;                   /* Socket */
    char*     multicastIP;            /* Arg: IP Multicast address */
    char*     multicastPort;          /* Arg: Server port */
    char*     sendString;             /* Arg: String to multicast */
    size_t    sendStringLen;          /* Length of string to multicast */
    struct addrinfo * multicastAddr;          /* Multicast address */
    struct addrinfo   hints          = { 0 }; /* Hints for name lookup */

    if ( argc != 4 )
    {
        fprintf(stderr, "Usage:  %s <Multicast Address> <Port> <Send String>\n", argv[0]);
        exit(10);
    }

    multicastIP   = argv[1];             /* First arg:   multicast IP address */
    multicastPort = argv[2];             /* Second arg:  multicast port */
    sendString    = argv[3];             /* Third arg:   String to multicast */
    sendStringLen = strlen(sendString);  /* Find length of sendString */

    /* Resolve destination address for multicast datagrams */
    hints.ai_family   = PF_INET6;
    hints.ai_socktype = SOCK_DGRAM;
    hints.ai_flags    = AI_NUMERICHOST;
    if (getaddrinfo(multicastIP, multicastPort, &hints, &multicastAddr) != 0) DieWithError("getaddrinfo() failed");

    /* Create socket for sending multicast datagrams */
    if ((sock = socket(multicastAddr->ai_family, multicastAddr->ai_socktype, 0)) == -1) DieWithError("socket() failed");

    int hops = 5;
    if (setsockopt(sock, IPPROTO_IPV6, IPV6_MULTICAST_HOPS, &hops, sizeof(hops)) != 0) DieWithError("setsockopt(MULTICAST_HOPS) failed");

    for (;;) /* Run forever */
    {
        int sendLen = sendto(sock, sendString, sendStringLen, 0, multicastAddr->ai_addr, multicastAddr->ai_addrlen);
        if (sendLen == sendStringLen )
        {
            printf("Sent [%s] (%i bytes) to %s, port %s\n", sendString, sendLen, multicastIP, multicastPort);
        }
        else
        {
            DieWithError("sendto() sent a different number of bytes than expected");
        }
        sleep(1); /* Multicast sendString in datagram to clients every second */
    }

    /* NOT REACHED */
    freeaddrinfo(multicastAddr);
    close(sock);
    return 0;
}
