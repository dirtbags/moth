/* multicast_server.c
 * Adapted from tmouse's IPv6 client/server example code
 * found at http://cboard.cprogramming.com/showthread.php?t=67469
 */

#include <netdb.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/time.h>
#include <unistd.h>

static void
DieWithError(const char* errorMessage)
{
  fprintf(stderr, "%s\n", errorMessage);
  exit(1);
}

int
main(int argc, char *argv[])
{
  int       sender, listener;	        /* Sockets */
  char*     multicastIP;                  /* Arg: IP Multicast address */
  char*     multicastPort;                /* Arg: Server port */
  char      token[100];
  struct    addrinfo *  multicastAddr;     /* Multicast address */
  struct    addrinfo    hints = { 0 };	/* Hints for name lookup */
  struct    timeval timeout = { 0 };

  if (argc != 3)
    {
      fprintf(stderr, "Usage:  %s ADDRESS PORT <TOKENFILE\n", argv[0]);
      exit(1);
    }

  multicastIP   = argv[1];             /* First arg:   multicast IP address */
  multicastPort = argv[2];             /* Second arg:  multicast port */

  if (NULL == fgets(token, sizeof(token), stdin)) {
    DieWithError("Unable to read token");
  }

  /* Resolve destination address for multicast datagrams */
  hints.ai_family   = PF_INET6;
  hints.ai_socktype = SOCK_DGRAM;
  hints.ai_flags    = AI_NUMERICHOST;

  if (getaddrinfo(multicastIP, multicastPort, &hints, &multicastAddr) != 0) {
    DieWithError("getaddrinfo() failed");
  }

  if (! ((multicastAddr->ai_family == PF_INET6) &&
         (multicastAddr->ai_addrlen == sizeof(struct sockaddr_in6)))) {
    DieWithError("Not IPv6");
  }

  /* Create socket for sending multicast datagrams */
  if ((sender = socket(multicastAddr->ai_family, multicastAddr->ai_socktype, 0)) == -1) {
    DieWithError("socket() failed");
  }

  /* Create socket for recieving multicast datagrams */
  if ((listener = socket(multicastAddr->ai_family, multicastAddr->ai_socktype, 0)) == -1) {
    DieWithError("socket() failed");
  }

  /* We need to go through a router, set hops to 5 */
  {
    int hops = 5;

    if (setsockopt(sender, IPPROTO_IPV6, IPV6_MULTICAST_HOPS, &hops, sizeof(hops)) != 0) {
      DieWithError("setsockopt(MULTICAST_HOPS) failed");
    }
  }

  /* Bind to the multicast port */
  if (bind(listener, multicastAddr->ai_addr, multicastAddr->ai_addrlen) != 0) {
    DieWithError("bind() failed");
  }


  /* Join the multicast group.  */
  {
    struct sockaddr_in6 *addr = (struct sockaddr_in6 *)(multicastAddr->ai_addr);
    struct ipv6_mreq multicastRequest;

    multicastRequest.ipv6mr_interface = addr->sin6_scope_id;
    memcpy(&multicastRequest.ipv6mr_multiaddr, &(addr->sin6_addr),
           sizeof(multicastRequest.ipv6mr_multiaddr));

    if (setsockopt(listener, IPPROTO_IPV6, IPV6_JOIN_GROUP,
                   (char*)&multicastRequest, sizeof(multicastRequest)) != 0) {
      DieWithError("setsockopt(IPV6_JOIN_GROUP) failed");
    }
  }

  for (;;) { /* Run forever */
    int       n;
    int       max_fd;
    fd_set    input;

    char      recvString[500];      /* Buffer for received string */
    int       recvStringLen;        /* Length of received string */

    char      sendString[] = "If anyone is out there, please say hello\n";
    size_t    sendStringLen = sizeof(sendString)-1;
    char      errorString[] = "Say what?\n";

    struct    sockaddr_in6  from;
    socklen_t fromlen = sizeof(from);

    FD_ZERO(&input);
    FD_SET(listener, &input);

    max_fd = listener + 1;

    if (timeout.tv_usec < 100) {
      ssize_t sendLen;

      timeout.tv_sec = 1;
      timeout.tv_usec = 0;

      sendLen = sendto(sender, sendString, sendStringLen, 0, multicastAddr->ai_addr,
                       multicastAddr->ai_addrlen);
      if (sendLen != sendStringLen) {
        DieWithError("sendto() sent a different number of bytes than expected");
      }
    }

    n = select(max_fd, &input, NULL, NULL, &timeout);

    /* See if there was an error */
    if (n < 0) {
      perror("select failed");
    } else if (FD_ISSET(listener, &input)) {
      recvStringLen = recvfrom(listener, recvString, sizeof(recvString) - 1, 0,
                               (struct sockaddr *)&from, &fromlen);
      /* Receive a single datagram from the server */
      if (recvStringLen < 0) {
        DieWithError("recvfrom() failed");
      }

      recvString[recvStringLen] = '\0';
      if (strcmp(recvString, "hello")==0) {
        sendto(listener, token, sizeof(sendString), 0, (struct sockaddr *)&from,
               fromlen);
      } else if (strcmp(recvString, sendString)!=0) {
        sendto(listener, errorString, sizeof(errorString), 0,
               (struct sockaddr *)&from, fromlen);
      }
    }
  }

  /* NOT REACHED */
  freeaddrinfo(multicastAddr);
  close(sender);
  close(listener);

  return 0;
}
