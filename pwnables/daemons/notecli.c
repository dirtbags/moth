#include <stdio.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <string.h>

int
open_connection(char *host, int port)
{
  struct sockaddr_in  addr;
  struct hostent     *h;
  int                 fd;

  if (! inet_aton(host, &(addr.sin_addr))) {
    if (!(h = gethostbyname(host))) {
      return -1;
    } else {
      memcpy(&(addr.sin_addr), h->h_addr, h->h_length);
    }
  }
  fd = socket(PF_INET, SOCK_STREAM, IPPROTO_TCP);
  if (fd == -1) {
    return -1;
  }
  addr.sin_family = AF_INET;
  addr.sin_port = htons(port);
  if (connect(fd, (struct sockaddr *)&addr, sizeof(addr))) {
    close(fd);
    return -1;
  }

  return fd;
}

void
clean(char *s)
{
  for (; *s; s++) {
    switch (*s) {
      case '/':
      case '.':
        *s = '_';
    }
  }
}

int
main(int argc, char *argv[])
{
  int     fd;
  char    s[4096];
  ssize_t len;

  if (4 != argc) {
    printf("Usage: %s HOST NOTE COMMAND\n", argv[0]);
    printf("\n");
    printf("  COMMAND must be 'r' (read) or 'w' (write).\n");
    printf("  For w, input is taken from stdin.\n");
    return 1;
  }

  fd = open_connection(argv[1], 4000);
  clean(argv[2]);
  switch (argv[3][0]) {
    case 'r':
      write(fd, "r", 1);
      write(fd, argv[2], strlen(argv[2]));
      write(fd, "\n", 1);
      do {
        len = read(fd, s, sizeof(s));
        write(1, s, len);
      } while (len);
      break;
    case 'w':
      write(fd, "w", 1);
      write(fd, argv[2], strlen(argv[2]));
      write(fd, "\n", 1);
      do {
        len = read(0, s, sizeof(s));
        write(fd, s, len);
      } while (len);
      break;
    default:
      printf("I don't understand that command.\n");
      break;
  }

  return 0;
}
