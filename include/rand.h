#ifndef __RAND_H__
#define __RAND_H__

#include <stdint.h>
#include <stddef.h>

void urandom(void *buf, size_t buflen);
int32_t rand32();
uint32_t randu32();

#endif /* __RAND_H__ */
