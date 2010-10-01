#include <sys/select.h>
#include <time.h>
#include <stdlib.h>
#include <stdio.h>
#include "obj.h"

#define NO_DEBUG
#define PID_MAX 32768
#define QSIZE 200
#define MSGS_PER_SEC 10

char const *token1 = "logger:token1";

int
randint(int max)
{
  return random() % max;
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
  time_t then = time(NULL);

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
#ifdef DEBUG
    time_t          now = then + 1;
#else
    time_t          now = time(NULL);
#endif
    int             i, max;

    /* Print messages */
    while ((msg = dequeue_message(now))) {
      char       ftime[80];
      struct tm *tm;

      tm = gmtime(&msg->when);
      if (! tm) {
        snprintf(ftime, sizeof(ftime), "%l", now);
      } else {
        strftime(ftime, sizeof(ftime), "%b %d %T", tm);
      }
      printf("%s loghost %s\n", ftime, msg->text);
    }
    fflush(stdout);

    /* Time for new tokens? */
#ifdef DEBUG
    then = now;
#else
    if (then + 60 <= now) {
      /* XXX: read in new tokens */
      then = now;
    }
#endif

    /* Make some messages */
    max = randint(MSGS_PER_SEC);

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
                     pid, token1);
            enqueue_messages(messages, 1);
          }
          /* Always follow this with a couple lines of fluff so it's
             not the last thing in a batch */
          max += 2;
          break;
        case 2:
          /* IMAP */
          if (-1 != get_many_messages(messages, 2)) {
            char const *u = user();

            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "imapd[%d]: Login: user=%s method=PLAIN",
                     pid, u);

            messages[1]->when = start + 2 + randint(60);
            snprintf(messages[1]->text, sizeof(messages[1]->text),
                     "imapd[%d]: Disconnected: Logged out");

            enqueue_messages(messages, 2);
          }
        case 3:
          /* IRC */
          if (-1 != get_many_messages(messages, 3)) {
            int connection = randint(512);
            int port       = randint(65536);

            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "ircd: Accepted connection %d from %d.%d.%d.%d:%d on socket %d.",
                     pid, connection,
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
          if (-1 != get_many_messages(messages, 8)) {
            char const *u    = user();
            int         o1   = randint(256);
            int         o2   = randint(256);
            int         o3   = randint(256);
            int         o4   = randint(256);
            long int    mid  = random();
            long int    mid2 = random();

            messages[0]->when = start;
            snprintf(messages[0]->text, sizeof(messages[0]->text),
                     "smtp/smtpd[%d]: connect from unknown[%d.%d.%d.%d]",
                     pid, o1, o2, o3, o4);

            messages[1]->when = messages[0]->when + randint(1);
            snprintf(messages[1]->text, sizeof(messages[1]->text),
                     "smtp/smtpd[%d]: %08X: client=unknown[%d.%d.%d.%d]",
                     pid, mid, o1, o2, o3, o4);

            messages[2]->when = messages[1]->when + 2 + randint(3);
            snprintf(messages[2]->text, sizeof(messages[2]->text),
                     "smtp/smtpd[%d]: disconnect from [%d.%d.%d.%d]",
                     pid, o1, o2, o3, o4);

            pid = (pid + 1 + randint(5)) % PID_MAX;
            messages[3]->when = messages[1]->when + 1 + randint(2);
            snprintf(messages[3]->text, sizeof(messages[3]->text),
                     "smtp/cleanup[%d]: %08X: message-id=<%08x@junkmail.spam>",
                     pid, mid, mid2);

            pid = (pid + 1 + randint(5)) % PID_MAX;
            messages[4]->when = messages[3]->when + randint(1);
            snprintf(messages[4]->text, sizeof(messages[4]->text),
                     "smtp/qmgr[%d]: %08X: from=<%s@junkmail.spam>, size=%d, nrcpt=1 (queue active)",
                     pid, mid, user(), randint(6000));

            messages[5]->when = messages[4]->when + 2 + randint(2);
            snprintf(messages[5]->text, sizeof(messages[5]->text),
                     "smtp/qmgr[%d]: %08X: removed",
                     pid, mid);

            messages[6]->when = messages[4]->when + randint(1);
            snprintf(messages[6]->text, sizeof(messages[6]->text),
                     "smtp/deliver(%s): msgid=<%08x@junkmail.spam>: saved to INBOX",
                     u, mid2);

            pid = (pid + 1 + randint(5)) % PID_MAX;
            messages[7]->when = messages[4]->when + randint(1);
            snprintf(messages[7]->text, sizeof(messages[7]->text),
                     "smtp/local[%d]: %08X: to <%s@dirtbags.net>, relay=local, dsn=2.0.0, status=sent (delivered to command /usr/bin/deliver)",
                     pid, mid, u);

            enqueue_messages(messages, 8);
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

#ifndef DEBUG
    sleep(1);
#endif
  }

  return 0;
}
