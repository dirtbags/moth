#include <stdio.h>
#include <syslog.h>

void
readntrim(char *s)
{
  gets(s);
  for (; *s; s++) {
    switch (*s) {
      case '\n':
      case '\r':
        *s = 0;
    }
  }
}

int
main(int argc, char *argv)
{
  int   cmd;
  char  line[4096];
  char  note[512];
  FILE *f = NULL;
  char   *peer = getenv("REMOTEADDR");

  openlog("in.noted", LOG_PID, LOG_USER);
  switch (getc(stdin)) {
    case EOF:
      return 0;
    case 'r':
      readntrim(note);
      if (peer) {
        syslog(LOG_INFO, "%s read %s", peer, note);
      }
      f = fopen(note, "r");
      while (fgets(line, sizeof(line), f)) {
        fputs(line, stdout);
      }
      fclose(f);
      break;
    case 'w':
      readntrim(note);
      if (peer) {
        syslog(LOG_INFO, "%s write %s", peer, note);
      }
      f = fopen(note, "w");
      while (gets(line)) {
        fputs(line, f);
      }
      fclose(f);
      break;
  }

  return 0;
}
