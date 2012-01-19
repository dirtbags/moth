#include <stdio.h>
#include <ctype.h>
#include <stdlib.h>
#include <string.h>

#define TRUE 1
#define FALSE 0

// finds the index of a character in the index
// returns index or -1
int
indexkey(char* key, char c) {
  int i;

  for(i=0; i < 25; i++) {
    if (key[i] == c) {
      //printf("'%d' -> %d\n", c, i);
      return i;
    }
  }
  return -1;
}

// makes sure everything is lowercase or a space
void
strtolower(char* s, int len) {
  int i;

  for(i = 0; i < len; i++) {
    s[i] = tolower(s[i]);
    if (s[i] < 'a' || s[i] > 'z') {
      s[i] = ' ';
    }
  }
}

//
// makes a key
char *
make_key(char* s, int len) {
  strtolower(s, len);
  char alph[] = "abcdefghijklmnopqrstuvwxyz";
  char* key = (char *) malloc(26 * sizeof(char));
  key[26] = '\0';
  int keylen = 0;
  int i;

  // initial dump
  for(i=0; i< len; i++) {
    if( s[i] != ' ' && alph[s[i]-97] != ' ' && s[i] != 'q') {
      key[keylen] = s[i];
      keylen++;
      alph[s[i]-97] = ' ';
    }
  }

  // add extra chars
  for (i=0; i < 27; i++) {
    if (alph[i] != ' ' && alph[i] != 'q') {
      key[keylen] = alph[i];
      keylen++;
      alph[i] = ' ';
    }
  }

  return key;
}


// double checks for duplicate chars in string
int
isdup(char* s, int len) {
  int i, j;

  for(i = 0; i < len; i++) {
    for(j = i+1; j < len; j++ ) {
      if (s[i] == s[j]) {
        return 1;
      }
    }
  }
  return 0;
}

// does the swapping of two characters
// assuming input is already sanitized
void
swapchar(char* key, char* plain) {
  int i0, i1;
  i0 = indexkey(key, plain[0]);
  i1 = indexkey(key, plain[1]);

  // will hit this with double null, or double x
  if (i0 == i1){
    // so pass
  // vertical case
  } else if (i0%5 == i1%5) {
    plain[0] = key[(i0+5)%25];
    plain[1] = key[(i1+5)%25];
  // horizontal case
  } else if (i0/5 == i1/5) {
    plain[0] = key[(i0/5)*5 + (i0+1)%5];
    plain[1] = key[(i1/5)*5 + (i1+1)%5];
  // diagonal case
  } else {
    int b0 = i0%5;
    int b1 = i1%5;
    int diff;
    if (b0 > b1) {
        diff = b0 - b1;
        plain[0] = key[i0-diff];
        plain[1] = key[i1+diff];
    } else {
        diff = b1 - b0;
        plain[0] = key[i0+diff];
        plain[1] = key[i1-diff];
    }
  }
  return;
}

void
printcrap(char* buf){
  printf("%c%c ", buf[0]-32, buf[1]-32);
}

void
run(char* key) {
  char buf[3];
  char tmp;
  int existing = FALSE;

  buf[2] = 0;

  while (TRUE) {
    // read some crap in
    tmp = getchar();
    if (tmp == 'q') {
        tmp = 'x';
    }
    if (tmp == EOF) {
      if(existing) {
        buf[1] = 'x';
        swapchar(key, buf);
        printcrap(buf);
        existing = FALSE;
      } else {
        return;
      }
    } else if (tmp == '\n') {
       if(existing) {
        buf[1] = 'x';
        swapchar(key, buf);
        printcrap(buf);
        printf("\n");
        fflush(stdout);
        existing = FALSE;
      } else {
        printf("\n");
        fflush(stdout);
      }
    } else if (91 <= tmp && tmp <= 122) {
      if (existing) {
        if (tmp == buf[0] && tmp != 'x') {
          buf[1] = 'x';
          swapchar(key, buf);
          printcrap(buf);
          buf[0] = tmp;
        } else {
          buf[1] = tmp;
          swapchar(key, buf);
          printcrap(buf);
          existing = FALSE;
        }
      } else {
        buf[0] = tmp;
        existing = TRUE;
      }
    } else {
      //printf("\nOnly [a-z\\n]\n");
      //fflush(stdout);
    }
  }
}


int
main() {
  // Unusual token, since it has to satisfy some strict requirements.
  char key[] = "netkutalbcdfgrisox";
  int len = strlen(key);
  char * ckey = make_key(key, len);

  // All I know about trigraphs is that the gcc manual says I don't want
  // to know about trigraphs.
  printf("The key is the token.  ???:????\?-???\?-????\n");
  fflush(stdout);
  run(ckey);
}
