#include <syslog.h>
#include <stdio.h>

int
main(int argc, char *argv)
{
  char    user[256];
  char    path[512];
  char   *data;
  FILE   *f;
  size_t  count;
  int     i;
  char   *peer = getenv("REMOTEADDR");

  openlog("in.fingerd", LOG_PID, LOG_USER);
  if (NULL == gets(user)) {
    return 0;
  }
  for (data = user; *data; data += 1) {
    if ('\r' == *data) {
      *data = 0;
    }
  }
  if (peer) {
    syslog(LOG_INFO, "%s requests %s", peer, user);
  }
  if (0 == user[0]) {
    printf("Nobody's home.\n");
    return 0;
  }

  sprintf(path, "/home/%s/.plan", user);
  f = fopen(path, "r");
  if (NULL == f) {
    printf("No such user.\n");
    return 0;
  }

  data = path;
  while (count = fread(data, sizeof(*data), 1, f)) {
    fwrite(data, count, 1, stdout);
  }
  return 0;
}
