#ifndef __XXTEA_H__
#define __XXTEA_H__

#include <stdint.h>

void tea_encode(uint32_t const key[4], uint32_t *buf, size_t buflen);
void tea_decode(uint32_t const key[4], uint32_t *buf, size_t buflen);

#endif
