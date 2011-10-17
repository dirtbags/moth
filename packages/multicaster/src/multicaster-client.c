/* multicast_client.c
 * Adopted from tmouse's client/server example code
 * found at http://cboard.cprogramming.com/showthread.php?t=67469
 */

#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <stdio.h>      /* for printf() and fprintf() */
#include <stdlib.h>     /* for atoi() and exit() */
#include <string.h>     /* for memset() */
#include <time.h>       /* for timestamps */
#include <unistd.h>

void DieWithError(const char* errorMessage)
{
    fprintf(stderr, "%s\n", errorMessage);
    exit(10);
}

int main(int argc, char* argv[])
{
    int     sock;                     /* Socket */
    char*      multicastIP;              /* Arg: IP Multicast Address */
    char*      multicastPort;            /* Arg: Port */
    struct addrinfo *  multicastAddr = {0};            /* Multicast Address */
    struct addrinfo *  localAddr;                /* Local address to bind to */
    struct addrinfo    hints          = { 0 };   /* Hints for name lookup */

    if ( argc != 3 )
    {
        fprintf(stderr,"Usage: %s <Multicast IP> <Multicast Port>\n", argv[0]);
        exit(10);
    }

    multicastIP   = argv[1];      /* First arg:  Multicast IP address */
    multicastPort = argv[2];      /* Second arg: Multicast port */

    /* Resolve the multicast group address */
    hints.ai_family = PF_INET6;
    hints.ai_flags  = AI_NUMERICHOST;
    if ( getaddrinfo(multicastIP, NULL, &hints, &multicastAddr) != 0 ) DieWithError("getaddrinfo() failed");

    /* Get a local address with the same family as our multicast group */
    hints.ai_family   = multicastAddr->ai_family;
    hints.ai_socktype = SOCK_DGRAM;
    hints.ai_flags    = AI_PASSIVE; /* Return an address we can bind to */
    if ( getaddrinfo(NULL, multicastPort, &hints, &localAddr) != 0 )
    {
        DieWithError("getaddrinfo() failed");
    }

    /* Create socket for receiving datagrams */
    if ( (sock = socket(localAddr->ai_family, localAddr->ai_socktype, 0)) == -1 )
    {
        DieWithError("socket() failed");
    }

    const int trueValue = 1;
    setsockopt(sock, SOL_SOCKET, SO_REUSEADDR, (const void *) &trueValue, sizeof(trueValue));
#ifdef __APPLE__
    setsockopt(sock, SOL_SOCKET, SO_REUSEPORT, (const void *) &trueValue, sizeof(trueValue));
#endif

    /* Bind to the multicast port */
    if ( bind(sock, localAddr->ai_addr, localAddr->ai_addrlen) != 0 )
    {
        DieWithError("bind() failed");
    }

    /* Join the multicast group.  */
    if ((multicastAddr->ai_family == PF_INET6)&&(multicastAddr->ai_addrlen == sizeof(struct sockaddr_in6)))
    {
      struct sockaddr_in6 *addr = (struct sockaddr_in6 *)(multicastAddr->ai_addr);
      struct ipv6_mreq multicastRequest;  /* Multicast address join structure */

      /* Specify the multicast group */
      memcpy(&multicastRequest.ipv6mr_multiaddr, &((struct sockaddr_in6*)(multicastAddr->ai_addr))->sin6_addr, sizeof(multicastRequest.ipv6mr_multiaddr));

      printf("scope_id: %d\n", addr->sin6_scope_id);

      /* Accept multicast from any interface */
      multicastRequest.ipv6mr_interface = addr->sin6_scope_id;

      /* Join the multicast address */
      if ( setsockopt(sock, IPPROTO_IPV6, IPV6_JOIN_GROUP, (char*) &multicastRequest, sizeof(multicastRequest)) != 0 ) DieWithError("setsockopt(IPV6_JOIN_GROUP) failed");
    }
    else DieWithError("Not IPv6");

    freeaddrinfo(localAddr);
    freeaddrinfo(multicastAddr);

    for (;;) /* Run forever */
    {
        char      recvString[500];      /* Buffer for received string */
        int       recvStringLen;        /* Length of received string */
	struct    sockaddr_in6  from;
	socklen_t fromlen = sizeof(from);
	char	  sendString[] = "Token: banana\n";
	char	  errorString[] = "That is not correct! Try again!\n";

        /* Receive a single datagram from the server */
        if ((recvStringLen = recvfrom(sock, recvString, sizeof(recvString) - 1, 0, (struct sockaddr *)&from, &fromlen)) < 0) DieWithError("recvfrom() failed");
        recvString[recvStringLen] = '\0';
	if(strcmp(recvString, "hello")==0) {
		printf("Correct!!\n");
	//	printf("Token: banana\n");
		sendto(sock, sendString, sizeof(sendString) - 1, 0, (struct sockaddr *)&from, fromlen);
	} else {
	//	printf("That isn't correct! Try again!\n");
		sendto(sock, errorString, sizeof(errorString) - 1, 0, (struct sockaddr *)&from, fromlen);
	}

        /* Print the received string */
        printf("Received string [%s]\n", recvString);
    }

    /* NOT REACHED */
    close(sock);
    exit(EXIT_SUCCESS);
}
