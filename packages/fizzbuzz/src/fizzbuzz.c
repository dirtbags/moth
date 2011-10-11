#include <stdio.h>
#include <stdlib.h>

/*
 * How this works:
 *
 * You have to provide this with the output of a fizzbuzz program to get
 * it to decode the token.
 *
 * Provide the encoded token on fd 3, and it will output the decode
 * provided correct input.  If you provided the decoded token, it will
 * encode it.  In other words, encode(x) = decode(x).
 *
 *
 * Here's a fizzbuzz program in bourne shell:
 *
 *   for i in $(seq 100); do
 *      if [ $(expr $i % 15) = 0 ]; then
 *          echo 'FizzBuzz'
 *      elif [ $(expr $i % 3) = 0 ]; then
 *          echo 'Fizz'
 *      elif [ $(expr $i % 5) = 0 ]; then
 *          echo 'Buzz'
 *      else
 *          echo $i
 *      fi
 *  done
 * 
 */

char craptable[] = {
  0x64, 0xd4, 0x11, 0x55, 0x50, 0x16, 0x61, 0x02,
  0xf7, 0xfd, 0x63, 0x36, 0xd9, 0xa6, 0xf2, 0x29,
  0xad, 0xfb, 0xed, 0x7a, 0x06, 0x91, 0xe7, 0x67,
  0x80, 0xb6, 0x53, 0x2c, 0x43, 0xf9, 0x3c, 0xf2,
  0x83, 0x5c, 0x25, 0xee, 0x21
};

int
main(int argc, char *argv[])
{
  int    i;
  char   token[100];
  size_t tokenlen;

  {
    FILE *tokenin = fdopen(3, "r");

    if (! tokenin) {
      printf("Somebody didn't read the instructions.\n");
      return 1;
    }
    
    tokenlen = fread(token, 1, sizeof(token), tokenin);
    fclose(tokenin);
  }
  

  for (i=1; i <= 100; i += 1) {
    char l[100];

    fgets(l, sizeof(l), stdin);

    if (0 == i % 15) {
      if (0 != strcmp("FizzBuzz\n", l)) break;
    } else if (0 == i % 3) {
      if (0 != strcmp("Fizz\n", l)) break;
    } else if (0 == i % 5) {
      if (0 != strcmp("Buzz\n", l)) break;
    } else {
      if (atoi(l) != i) break;
    }

    token[i % tokenlen] ^= i;
    token[i % tokenlen] ^= craptable[i % sizeof(craptable)];
  }

  if (101 == i) {
    fwrite(token, tokenlen, 1, stdout);
  } else {
    printf("Next time, FizzBuzz me.\n");
  }

  return 0;
}
