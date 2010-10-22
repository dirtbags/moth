/*
------------------------------------------------------------------------------
rand.h: definitions for a random number generator
By Bob Jenkins, 1996, Public Domain
MODIFIED:
  960327: Creation (addition of randinit, really)
  970719: use context, not global variables, for internal state
  980324: renamed seed to flag
  980605: recommend RANDSIZL=4 for noncryptography.
  010626: note this is public domain
  101005: update to C99 (neale@lanl.gov)
------------------------------------------------------------------------------
*/

#ifndef __ISAAC_H__
#define __ISAAC_H__

#include <stdint.h>

#define RANDSIZL   (8)
#define RANDSIZ    (1<<RANDSIZL)

/* context of random number generator */
struct randctx {
    uint32_t randcnt;
    uint32_t randrsl[RANDSIZ];
    uint32_t randmem[RANDSIZ];
    uint32_t randa;
    uint32_t randb;
    uint32_t randc;
};

/*
------------------------------------------------------------------------------
 If (flag==TRUE), then use the contents of randrsl[0..RANDSIZ-1] as the seed.
------------------------------------------------------------------------------
*/
void randinit(struct randctx *ctx, uint_fast8_t flag);

void isaac(struct randctx *ctx);

/*
------------------------------------------------------------------------------
 Call rand(/o_ randctx *r _o/) to retrieve a single 32-bit random value
------------------------------------------------------------------------------
*/
#define rand32(r) \
   (!(r)->randcnt-- ? \
     (isaac(r), (r)->randcnt=RANDSIZ-1, (r)->randrsl[(r)->randcnt]) : \
     (r)->randrsl[(r)->randcnt])

#endif				/* RAND */


#endif /* __ISAAC_H__ */
