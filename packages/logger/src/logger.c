#include <sys/select.h>
#include <time.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include "token.h"

#define PID_MAX 32768
#define QSIZE 200
#define MSGS_PER_SEC_MIN 10
#define MSGS_PER_SEC_MAX 40

const uint8_t key[] = {0x99, 0xeb, 0xc0, 0xce,
                       0xe0, 0xc9, 0xed, 0x5b,
                       0xbd, 0xc8, 0xb5, 0xfd,
                       0xdd, 0x0b, 0x03, 0x10};

/* Storage space for tokens */
char token[3][TOKEN_MAX];

void
read_tokens()
{
  int     i;
  ssize_t len;
  char    name[40];

  for (i = 0; i < sizeof(token)/sizeof(*token); i += 1) {
    /* This can't grow beyond 40.  Think about it. */
    sprintf(name, "logger%d", i);

    len = read_token(name, key, sizeof(key), token[i], sizeof(token[i]));
    if ((-1 == len) || (len >= sizeof(token[i]))) abort();
    token[i][len] = '\0';
  }
}


/*
 * Base 64 (GPL: see COPYING)
 */

/* C89 compliant way to cast 'char' to 'unsigned char'. */
static inline unsigned char
to_uchar (char ch)
{
  return ch;
}

/* Base64 encode IN array of size INLEN into OUT array of size OUTLEN.
   If OUTLEN is less than BASE64_LENGTH(INLEN), write as many bytes as
   possible.  If OUTLEN is larger than BASE64_LENGTH(INLEN), also zero
   terminate the output buffer. */
void
base64_encode (const char *in, size_t inlen,
               char *out, size_t outlen)
{
  static const char b64str[64] =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

  while (inlen && outlen) {
    *out++ = b64str[(to_uchar(in[0]) >> 2) & 0x3f];
    if (!--outlen)
      break;
    *out++ = b64str[((to_uchar(in[0]) << 4)
                     + (--inlen ? to_uchar(in[1]) >> 4 : 0))
                    & 0x3f];
    if (!--outlen)
      break;
    *out++ = (inlen
              ? b64str[((to_uchar(in[1]) << 2)
                        + (--inlen ? to_uchar(in[2]) >> 6 : 0))
                       & 0x3f]
              : '=');
    if (!--outlen)
      break;
    *out++ = inlen ? b64str[to_uchar(in[2]) & 0x3f] : '=';
    if (!--outlen)
      break;
    if (inlen)
      inlen--;
    if (inlen)
      in += 3;
  }

  if (outlen)
    *out = '\0';
}


/*
 * Bubble Babble
 */
char const consonants[] = "bcdfghklmnprstvz";
char const vowels[]     = "aeiouy";

#define bubblebabble_len(n) (6*(((n)/2)+1))

/** Compute bubble babble for input buffer.
 *
 * The generated output will be of length 6*((inlen/2)+1), including the
 * trailing NULL.
 *
 * Test vectors:
 *     `' (empty string) `xexax'
 *     `1234567890'      `xesef-disof-gytuf-katof-movif-baxux'
 *     `Pineapple'       `xigak-nyryk-humil-bosek-sonax'
 */
void
bubblebabble(unsigned char *out,
             unsigned char const *in,
             const size_t inlen)
{
  size_t pos  = 0;
  int    seed = 1;
  size_t i    = 0;

  out[pos++] = 'x';
  while (1) {
    unsigned char c;

    if (i == inlen) {
      out[pos++] = vowels[seed % 6];
      out[pos++] = 'x';
      out[pos++] = vowels[seed / 6];
      break;
    }

    c = in[i++];
    out[pos++] = vowels[(((c >> 6) & 3) + seed) % 6];
    out[pos++] = consonants[(c >> 2) & 15];
    out[pos++] = vowels[((c & 3) + (seed / 6)) % 6];
    if (i == inlen) {
      break;
    }
    seed = ((seed * 5) + (c * 7) + in[i]) % 36;

    c = in[i++];
    out[pos++] = consonants[(c >> 4) & 15];
    out[pos++] = '-';
    out[pos++] = consonants[c & 15];
  }

  out[pos++] = 'x';
  out[pos] = '\0';
}



int
randint(int max)
{
  return random() % max;
}

#define itokenlen 5

char const *
bogus_token()
{
  static char   token[TOKEN_MAX];
  unsigned char crap[itokenlen];
  unsigned char digest[bubblebabble_len(itokenlen)];
  int           i;

  for (i = 0; i < sizeof(crap); i += 1 ) {
    crap[i] = (unsigned char)randint(256);
  }
  bubblebabble(digest, (unsigned char *)&crap, itokenlen);
  snprintf(token, sizeof(token), "bogus:%s", digest);
  token[sizeof(token) - 1] = '\0';

  return token;
}

#define choice(a) (a[randint(sizeof(a)/sizeof(*a))])

char const *users[] = {"alice", "bob", "carol", "dave",
                       "eve", "fran", "gordon",
                       "isaac", "justin", "mallory",
                       "oscar", "pat", "steve",
                       "trent", "vanna", "walter", "zoe"};


char const *
user()
{
  return choice(users);
}

char const *filenames[] = {"about", "request", "page", "buttons",
                           "images", "overview"};
char const *extensions[] = {"html", "htm", "jpg", "png", "css", "cgi"};
char const *fields[] = {"q", "s", "search", "id", "req", "oid", "pmt",
                        "u", "page", "xxnp", "stat", "jk", "ttb",
                        "access", "domain", "needle", "service", "client"};
char const *values[] = {"1", "turnip", "chupacabra", "58", "identify",
                        "parthenon", "jellyfish", "pullman", "auth",
                        "xa4Jmwl", "cornmeal", "ribbon", "49299248",
                        "javaWidget", "crashdump", "priority",
                        "blogosphere"};

char const *
url()
{
  static char url[200];
  int         i, parts;

  strcpy(url, "/");

  parts = randint(4);
  for (i = 0; i < parts; i += 1) {
    if (i > 0) {
      strcat(url, "/");
    }
    strcat(url, choice(filenames));
  }

  if (randint(5) > 1) {
    if (i > 0) {
      strcat(url, ".");
      strcat(url, choice(extensions));
    }
  } else {
    parts = randint(8) + 1;
    for (i = 0; i < parts; i += 1) {
      if (0 == i) {
        strcat(url, "?");
      } else {
        strcat(url, "&");
      }
      strcat(url, choice(fields));
      strcat(url, "=");
      strcat(url, choice(values));
    }
  }

  return url;
}


struct message {
  time_t          when;
  char            text[300];
  struct message *next;
};

/* Allocate some messages */
struct message heap[QSIZE];

struct message *pool;
struct message *queue;

struct message *
get_message()
{
  struct message *ret = pool;

  if (pool) {
    pool = pool->next;
  }

  return ret;
}

void
free_message(struct message *msg)
{
  if (msg) {
    msg->next = pool;
    pool = msg;
  }
}

/* Either get count messages, or don't get any at all. */
int
get_many_messages(struct message **msgs, size_t count)
{
  int i;

  for (i = 0; i < count; i += 1) {
    msgs[i] = get_message();
  }

  if (NULL == msgs[i-1]) {
    for (i = 0; i < count; i += 1) {
      free_message(msgs[i]);
    }
    return -1;
  }

  return 0;
}

void
enqueue_message(struct message *msg)
{
  struct message *cur;

  /* In some cases, we want msg to be at the head */
  if ((NULL == queue) || (queue->when > msg->when)) {
    msg->next = queue;
    queue = msg;
    return;
  }

  /* Find where to stick it */
  for (cur = queue; NULL != cur->next; cur = cur->next) {
    if (cur->next->when > msg->when) break;
  }

  /* Insert it after cur */
  msg->next = cur->next;
  cur->next = msg;
}

void
enqueue_messages(struct message **msgs, size_t count)
{
  int i;

  for (i = 0; i < count; i += 1) {
    enqueue_message(msgs[i]);
  }
}

struct message *
dequeue_message(time_t now)
{
  if ((NULL != queue) && (queue->when <= now)) {
    struct message *ret = queue;

    queue = queue->next;
    free_message(ret);
    return ret;
  }

  return NULL;
}

int
main(int argc, char *argv[])
{
  int    i;
  int    pid  = 52;
  time_t then = time(NULL) - 100; /* Assure we get new tokens right away */

  /* Seed RNG */
  srandom(then);

  /* Initialize free messages */
  {
    pool = &(heap[0]);
    for (i = 0; i < QSIZE - 1; i += 1) {
      heap[i].next = &(heap[i+1]);
    }
    heap[i].next = NULL;
  }

  /* Now let's make some crap! */
  while (! feof(stdout)) {
    struct message *msg;
    time_t          now = time(NULL);
    int             i, max;

    /* Print messages */
    while ((msg = dequeue_message(now))) {
      char       ftime[80];
      struct tm *tm;

      tm = gmtime(&msg->when);
      if (! tm) {
        snprintf(ftime, sizeof(ftime), "%ld", now);
      } else {
        strftime(ftime, sizeof(ftime), "%b %d %T", tm);
      }
      printf("%s loghost %s\n", ftime, msg->text);
    }
    fflush(stdout);

    /* Time for new tokens? */
    if (then + 60 <= now) {
      read_tokens();
      then = now;
    }

    /* Make some messages */
    max = MSGS_PER_SEC_MIN + randint(MSGS_PER_SEC_MAX - MSGS_PER_SEC_MIN);

    for (i = 0; i < max; i += 1) {
      time_t          start = now + 1;
      struct message *messages[10];

      /* Increment the PID */
      pid = (pid + 1 + randint(20)) % PID_MAX;

      switch (randint(90)) {
        case 0:
          /* Internal diagnostic! */
          if (-1 != get_many_messages(messages, 1)) {
            int             queued, pooled;
            struct message *msg;

            for (pooled = 0, msg = pool;
                 msg;
                 msg = msg->next, pooled += 1);
            /* Start at one because of this message */
            for (queued = 1, msg = queue;
                 msg;
                 msg = msg->next, queued += 1);

            messages[0]->when = now;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "DEBUG: %d in pool, %d in queue (%d total)",
                     pooled, queued, pooled + queued);
            enqueue_messages(messages, 1);
          }
        case 1:
          /* Lame-o "token" service */
          if (-1 != get_many_messages(messages, 1)) {
            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "tokenserv[%d]: token is %s",
                     pid, token[0]);
            enqueue_messages(messages, 1);
          }
          /* Always follow this with a couple lines of fluff so it's
             not the last thing in a batch */
          max += 2;
          break;
        case 2:
          /* IMAP */
          {
            char const *mytoken;
            char const *u;
            char btoken[TOKEN_MAX * 2];

            if (randint(5) == 0) {
              mytoken = token[1];
              u = "token";
            } else {
              mytoken = bogus_token();
              u = user();
            }
            base64_encode(mytoken, strlen(mytoken), btoken, sizeof(btoken));

            if (-1 != get_many_messages(messages, 2)) {
              const int offset=15;

              messages[0]->when = start;
              snprintf(messages[0]->text, sizeof(messages[0]->text),
                       "imapd[%d]: Login: user=%s method=PLAIN token1=%.*s",
                       pid, u, offset, btoken);

              messages[1]->when = start + 4 + randint(60);
              snprintf(messages[1]->text, sizeof(messages[1]->text),
                       "imapd[%d]: Disconnected: Logged out token2=%s",
                       pid, btoken + offset);

              enqueue_messages(messages, 2);
            }
          }
        case 3:
          /* IRC */
          if (-1 != get_many_messages(messages, 3)) {
            int connection = randint(512);
            int port       = randint(65536);

            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "ircd: Accepted connection %d from %d.%d.%d.%d:%d on socket %d.",
                     connection,
                     randint(256), randint(256),
                     randint(256), randint(256),
                     port,
                     randint(256));

            messages[1]->when = start + randint(5);
            snprintf(messages[1]->text, sizeof(messages[1]->text),
                     "ircd: User \"%s!~%s@dirtbags.net\" registered (connection %d).",
                     user(), user(), connection);


            messages[2]->when = messages[1]->when + randint(600);
            snprintf(messages[2]->text, sizeof(messages[2]->text),
                     "ircd: Shutting down connection %d (Got QUIT command.) with dirtbags.net:%d.",
                     connection, port);

            enqueue_messages(messages, 3);
          }
          break;
        case 4:
          /* cron */
          if (-1 != get_many_messages(messages, 1)) {
            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "/USR/SBIN/CRON[%d]: (root) CMD (   /opt/bloatware/cleanup.sh )",
                     pid);
            enqueue_messages(messages, 1);
          }
          break;
        case 5:
          /* sudo */
          if (-1 != get_many_messages(messages, 1)) {
            char const *u = user();

            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "sudo: %12s : TTY=pts/%d ; PWD=/home/%s ; USER=root; COMMAND=/usr/bin/less /var/log/syslog",
                     u, randint(12), u);
            enqueue_messages(messages, 1);
          }
          break;
        case 6 ... 20:
          /* SMTP */
          {
            char const *mytoken;
            size_t      tokenlen;
            char const *host;
            size_t      hostlen;
            char const *from;
            size_t      fromlen;
            char const *to;
            int         is_token;

            if (randint(10) == 0) {
              is_token = 1;
              mytoken = token[2];
            } else {
              is_token = 0;
              mytoken = bogus_token();
            }

            tokenlen = strlen(mytoken);
            host = mytoken;
            hostlen = tokenlen/3;
            from = mytoken + hostlen;
            fromlen = tokenlen/3;
            to = mytoken + hostlen + fromlen;

            if (-1 != get_many_messages(messages, 8)) {
              int      o1   = randint(256);
              int      o2   = randint(256);
              int      o3   = randint(256);
              int      o4   = randint(256);
              long int mid  = random();
              long int mid2 = random();

              messages[0]->when = start;
              snprintf(messages[0]->text, sizeof(messages[0]->text),
                       "smtp/smtpd[%d]: connect from %.*s[%d.%d.%d.%d]",
                       pid, hostlen, host, o1, o2, o3, o4);

              messages[1]->when = messages[0]->when + randint(1);
              snprintf(messages[1]->text, sizeof(messages[1]->text),
                       "smtp/smtpd[%d]: %08lX: client=%.*s[%d.%d.%d.%d]",
                       pid, mid, hostlen, host, o1, o2, o3, o4);

              messages[2]->when = messages[1]->when + 2 + randint(3);
              snprintf(messages[2]->text, sizeof(messages[2]->text),
                       "smtp/smtpd[%d]: disconnect from [%d.%d.%d.%d]",
                       pid, o1, o2, o3, o4);

              pid = (pid + 1 + randint(5)) % PID_MAX;
              messages[3]->when = messages[1]->when + 1 + randint(2);
              snprintf(messages[3]->text, sizeof(messages[3]->text),
                       "smtp/cleanup[%d]: %08lX: message-id=<%08lx@junkmail.spam>",
                       pid, mid, mid2);

              pid = (pid + 1 + randint(5)) % PID_MAX;
              messages[4]->when = messages[3]->when + randint(1);
              snprintf(messages[4]->text, sizeof(messages[4]->text),
                       "smtp/qmgr[%d]: %08lX: from=<%.*s@junkmail.spam>, size=%d, nrcpt=1 (queue active)",
                       pid, mid, fromlen, from, randint(6000));

              messages[5]->when = messages[4]->when + 2 + randint(2);
              snprintf(messages[5]->text, sizeof(messages[5]->text),
                       "smtp/qmgr[%d]: %08lX: removed",
                       pid, mid);

              messages[6]->when = messages[4]->when + randint(1);
              snprintf(messages[6]->text, sizeof(messages[6]->text),
                       "smtp/deliver(%s): msgid=<%08lx@junkmail.spam>: saved to INBOX",
                       to, mid2);

              pid = (pid + 1 + randint(5)) % PID_MAX;
              messages[7]->when = messages[4]->when + randint(1);
              snprintf(messages[7]->text, sizeof(messages[7]->text),
                       "smtp/local[%d]: %08lX: to <%s@dirtbags.net>, relay=local, dsn=2.0.0, status=sent (delivered to command /usr/bin/deliver)",
                       pid, mid, to);

              enqueue_messages(messages, 8);
            }
          }
          break;
        case 21 ... 30:
          /* ssh */
          break;
        default:
          /* HTTP */
          if (-1 != get_many_messages(messages, 1)) {
            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "httpd[%d]: %d.%d.%d.%d\t-\tdirtbags.net\t80\tGET\t%s\t-\tHTTP/1.1\t200\t%d\t-\tMozilla/5.0",
                     pid,
                     randint(256), randint(256),
                     randint(256), randint(256),
                     url(), randint(4000) + 378);
            enqueue_messages(messages, 1);
          }
          break;
      }
    }

    sleep(1);
  }

  return 0;
}
