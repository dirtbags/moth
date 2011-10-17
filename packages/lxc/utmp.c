/* Detecting runlevels from utmp is straight up bullshit, you.

   1. runit doesn't have run levels
   2. dbtl doesn't write utmp
   3. even if it did, it doesn't have the glibc functions this code
      wants
*/
int lxc_utmp_mainloop_add(struct lxc_epoll_descr *descr,
			  struct lxc_handler *handler) {
  return 0;
}
