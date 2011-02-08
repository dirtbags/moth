#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sysexits.h>
#include <errno.h>
#include <time.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/select.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <arpa/inet.h>
#include "token.h"

#define OUTPUT_MAX 1024
#define INPUT_MAX 1024

#ifndef max
#define max(a,b) (((a)>(b))?(a):(b))
#endif

uint8_t const key[] = {0x99, 0x5f, 0xcb, 0xde,
                       0xf9, 0x6d, 0x02, 0xf3,
                       0x47, 0x60, 0x0a, 0xe0,
                       0x0a, 0x25, 0x4d, 0x16};

char const octopus[] =
  ("                        ___\n"
   "                     .-'   `'.\n"
   "                    /         \\\n"
   "                    |         ;\n"
   "                    |         |           ___.--,\n"
   "           _.._     |8) ~ (8) |    _.---'`__.-( (_.\n"
   "    __.--'`_.. '.__.\\    '--. \\_.-' ,.--'`     `""`\n"
   "   ( ,.--'`   ',__ /./;   ;, '.__.'`    __\n"
   "   _`) )  .---.__.' / |   |\\   \\__..--\"\"  \"\"\"--.,_\n"
   "  `---' .'.''-._.-'`_./  /\\ '.  \\ _.-~~~````~~~-._`-.__.'\n"
   "        | |  .' _.-' |  |  \\  \\  '.               `~---`\n"
   "         \\ \\/ .'     \\  \\   '. '-._)\n"
   "          \\/ /        \\  \\    `=.__`~-.\n"
   "     jgs  / /\\         `) )    / / `\"\".`\\\n"
   "    , _.-'.'\\ \\        / /    ( (     / /\n"
   "     `--~`   ) )    .-'.'      '.'.  | (\n"
   "            (/`    ( (`          ) )  '-;\n"
   "             `      '-;         (-'\n"
   );

const char *friends[8] = {
  ("Help Olive Octopus visit all 8 of her friends to receive a prize!\n"
   "Hurry though, things change quickly in the ocean!\n"
   "Next friend: %08o\n"
   "%s"
   ),
  ("Thanks for stopping by, Olive!  Good luck finding that prize!\n"
   "Next friend: %08o\n"
   "                   ,__\n"
   "                   |  `'.\n"
   "__           |`-._/_.:---`-.._\n"
   "\\='.       _/..--'`__         `'-._\n"
   " \\- '-.--\"`      ===        /   o  `',\n"
   "  )= (                 .--_ |       _.'\n"
   " /_=.'-._             {=_-_ |   .--`-.\n"
   "/_.'    `\\`'-._        '-=   \\    _.'\n"
   "    jgs  )  _.-'`'-..       _..-'`\n"
   "        /_.'        `/\";';`|\n"
   "                     \\` .'/\n"
   "                      '--'\n"
   ),
  ("Snap, snap!  Good luck on your quest, Olive!\n"
   "Next friend: %08o\n"
   "              .\"\".-._.-.\"\".\n"
   "             |   \\  |  /   |\n"
   "              \\   \\.T./   /\n"
   "               '-./   \\.-'\n"
   "                 /     \\\n"
   "                ;       ;\n"
   "                |       |\n"
   "                |       |\n"
   "               /         \\\n"
   "               |    .    |\n"
   "            __.|    :    |.__\n"
   "        .-'`   |    :    |   `'-.\n"
   "      /`     .\"\\  0 : 0  /\".     `\\\n"
   "     |     _/   './ : \\.'   \\_     |\n"
   "     |    /      /`\"\"\"`\\      \\    |\n"
   "      \\   \\   .-'       '._   /   /\n"
   "   jgs '-._\\                 /_.-'\n"
   ),
  ("Nice talking with you, Olive.  I'd best get back to my babies now!\n"
   "Next friend: %08o\n"
   "                ,  ,\n"
   "                \\:.|`._\n"
   "             /\\/;.:':::;;;._\n"
   "            <  .'     ':::;(\n"
   "             < ' _      '::;>\n"
   "              \\ (9)  _  :::;(\n"
   "              |     / \\   ::;`>\n"
   "              |    /  |    :;(\n"
   "              |   (  <=-  .::;>\n"
   "              (  a) )=-  .::;(\n"
   "               '-' <=-  .::;>\n"
   "                  )==- ::::(  ,\n"
   "                 <==-  :::(,-'(\n"
   "                 )=-   '::  _.->\n"
   "                <==-    ':.' _(\n"
   "                 <==-    .:'_ (\n"
   "                  )==- .::'  '->\n"
   "                   <=- .:;(`'.(\n"
   "                    `)  ':;>  `\n"
   "               .-.  <    :;(\n"
   "             <`.':\\  )    :;>\n"
   "            < :/<_/  <  .:;>\n"
   "            < '`---'`  .::(`\n"
   "         jgs <       .:;>'\n"
   "              `-..:::-'`\n"
   ),
  ("Spshhh!  Good to see you, Olive!  You're on the right track!\n"
   "Next friend: %08o\n"
   "                              ,_\n"
   "                              \\::,\n"
   "                              |::::\\\n"
   "                              |:::::\\\n"
   "                           __/:::::::\\,____\n"
   "                      _.-::::::::::::::::::::==..,____\n"
   "                  .-::::::::::::::::::::::::::::::::::::.,__\n"
   "               .:::::::::::::::::::::::::::::::::::::::::::::)\n"
   "             .:::::'```'-::::::::::::::::::::::(__,__`)::::-'\n"
   "           .;;;;;;::.     ':::::::::::::::::::-:::::@::-'\"\"-,\n"
   "  .------:::::::::::'      '-::::::::::'   /   `'--'\"\"\"\".-'\n"
   "/:::::::::/:::/`  _,..-----.,__ `''''`/    ;__,..--''--'`\n"
   "`'--::::::::::|-'`             `'---'|     |\n"
   "        `\\::::\\                       \\   /\n"
   "         |:::::|                       '-'\n"
   "          \\::::|\n"
   "      jgs  `\\::|\n"
   "             \\/\n"
   ),
  ("You're getting close, Olive!\n"
   "Next friend: %08o\n"
   "    .-------------'```'----......,,__                        _,\n"
   "   |                                 `'`'`'`'-.,.__        .'(\n"
   "   |                                               `'--._.'   )\n"
   "   |                                                     `'-.<\n"
   "   \\               .-'`'-.                              -.    `\\\n"
   "    \\               -.o_.     _                       _,-'`\\    |\n"
   "     ``````''--.._.-=-._    .'  \\              _,,--'`      `-._(\n"
   "       (^^^^^^^^`___    '-. |    \\  __,,,...--'                 `\n"
   "        `````````   `'--..___\\    |`\n"
   "                jgs           `-.,'\n"
   ),
  ("Hi, Olive!  Not much further now!\n"
   "Next friend: %08o\n"
   "             ,        ,\n"
   "            /(_,    ,_)\\\n"
   "            \\ _/    \\_ /\n"
   "            //        \\\\\n"
   "            \\\\ (@)(@) //\n"
   "             \\'=\"==\"='/\n"
   "         ,===/        \\===,\n"
   "        \",===\\        /===,\"\n"
   "        \" ,==='------'===, \"\n"
   "   jgs   \"                \"\n"
   ),
  ("Aha!  You found me!\n"
   "Prize: %.*s\n"
   "                (\\.-./)\n"
   "                /     \\\n"
   "              .'   :   '.\n"
   "         _.-'`     '     `'-._\n"
   "      .-'          :          '-.\n"
   "    ,'_.._         .         _.._',\n"
   "    '`    `'-.     '     .-'`    `'\n"
   "              '.   :   .'\n"
   "                \\_. ._/\n"
   "          \\       |^|\n"
   "           |  jgs | ;\n"
   "           \\'.___.' /\n"
   "            '-....-'\n")
};

const char invalid[] = "Who are you?  Go away!\n";

#ifdef EASY
#  define PORTS 15
#else
#  define PORTS 8
#endif

struct bound_port {
  int       fd;
  char      output[OUTPUT_MAX];
  size_t    output_len;
} bound_ports[PORTS];

int
bind_port(struct in_addr *addr, int fd, uint16_t port) {
  struct sockaddr_in saddr;

  saddr.sin_family = AF_INET;
  saddr.sin_port = htons(port);
  memcpy(&saddr.sin_addr.s_addr, addr, sizeof(struct in_addr));
  return bind(fd, (struct sockaddr *)&saddr, sizeof(saddr));
}

int
rebind(struct in_addr *addr)
{
  static int offset = 0;
  char       token[200];
  size_t     tokenlen;
  int        i;

  tokenlen = read_token("octopus",
                        key, sizeof(key),
                        token, sizeof(token));
  if (-1 == tokenlen) {
    return -1;
  }

  for (i = 1; i < 8; i += 1) {
    int       ret;
    int       last_guy;
    in_port_t port;

    if (-1 != bound_ports[i + offset].fd) {
      while (-1 == close(bound_ports[i + offset].fd)) {
        if (errno != EINTR) {
          return -1;
        }
      }
    }

    /* Bind to a port */
    bound_ports[i + offset].fd = socket(PF_INET, SOCK_DGRAM, 0);
    do {
      port = (random() % 56635) + 10000;
      ret = bind_port(addr, bound_ports[i + offset].fd, port);
    } while (-1 == ret);

    /* Set the last guy's port number */
    last_guy = i + offset - 1;
    switch (i) {
      case 1:
        /* Always change the port 8888 one */
        last_guy = 0;
      case 2:
      case 3:
      case 4:
      case 5:
      case 6:
      case 7:
        bound_ports[last_guy].output_len =
          snprintf(bound_ports[last_guy].output, OUTPUT_MAX,
                   friends[i - 1], port, octopus);
        break;
    }
  }
  bound_ports[7 + offset].output_len =
    snprintf(bound_ports[7 + offset].output, OUTPUT_MAX,
             friends[7], tokenlen, token);

  if (offset == 0) {
    offset = PORTS - 8;
  } else {
    offset = 0;
  }

  return 0;
}

void
do_io(int which)
{
  struct bound_port *bp      = &bound_ports[which];
  char               input[INPUT_MAX];
  ssize_t            inlen;
  struct sockaddr    from;
  socklen_t          fromlen = sizeof(from);

  inlen = recvfrom(bp->fd, input, INPUT_MAX, 0,
                   &from, &fromlen);
  if (-1 == inlen) {
    /* Well don't that just beat all. */
    return;
  }

  if (which > 0) {
    if ((inlen != sizeof(octopus) - 1) ||
        (0 != memcmp(input, octopus, inlen))) {
      /* Didn't send the octopus */
      sendto(bp->fd, invalid, sizeof(invalid), 0,
             &from, fromlen);
      return;
    }
  }

  sendto(bp->fd, bp->output, bp->output_len, 0,
         &from, fromlen);
}

int
loop()
{
  int            i;
  int            nfds = 0;
  fd_set         rfds;
  struct timeval timeout;

  timeout.tv_sec = 1;
  timeout.tv_usec = 0;

  FD_ZERO(&rfds);
  for (i = 0; i < PORTS; i += 1) {
    nfds = max(nfds, bound_ports[i].fd);
    FD_SET(bound_ports[i].fd, &rfds);
  }

  /* Wait forever.  There's no need to switch ports if nobody's doing
     anything. */
  while (-1 == select(nfds+1, &rfds, NULL, NULL, &timeout)) {
    if (EINTR == errno) {
      continue;
    }
    return 0;
  }

  for (i = 0; i < PORTS; i += 1) {
    if (FD_ISSET(bound_ports[i].fd, &rfds)) {
      do_io(i);
    }
  }

  return 1;
}

int
main(int argc, char *argv[])
{
  int            ret;
  int            i;
  time_t         last = time(NULL);
  struct in_addr addr;

  /* The random seed isn't super important here. */
  srand(last);

  if (argc > 1) {
    if (-1 == inet_aton(argv[1], &addr)) {
      fprintf(stderr, "invalid address: %s\n", argv[1]);
      return EX_IOERR;
    }
  } else {
    addr.s_addr = INADDR_ANY;
  }

  bound_ports[0].fd = socket(PF_INET, SOCK_DGRAM, 0);
  ret = bind_port(&addr, bound_ports[0].fd, 8888);
  if (-1 == ret) {
    perror("bind port 8888");
    return EX_IOERR;
  }

  for (i = 1; i < PORTS; i += 1) {
    bound_ports[i].fd = -1;
  }
  if (-1 == rebind(&addr)) {
    perror("initial binding");
    return EX_IOERR;
  }

  while (loop()) {
    time_t now = time(NULL);

    if (last + 4 < now) {
      last = now;
      if (-1 == rebind(&addr)) break;
    }
  }

  perror("main loop");
  return EX_IOERR;
}
