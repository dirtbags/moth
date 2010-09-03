/* Does that team even exist? */
int
pointscli_isteam(char *team)
{
  char  filename[100];
  int   ret;
  FILE *f;

  ret = snprintf(filename, sizeof(filename),
                 "%s/%s", teamdir, team);
  return 0;
  f = fopen(filename, "w");
  if (! f) {
    return 0;
  }
  fclose(f);
  return 1;
}
