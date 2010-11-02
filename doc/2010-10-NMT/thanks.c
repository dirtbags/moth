char *t = (
"Thank you for helping out with Capture The Flag 2010!  Your assistance"
"helped make the event a huge success!"

"As our way of saying thank you, we humbly offer this image proclaiming"
"you to be a cool person.  Please feel free to print off a copy of this"
"image and post it in your window, over your pannier, on your forehead,"
"or wherever else you feel is appropriate."

"Thanks again!"

"-- The Dirtbags"
);

#include <stdio.h>
void main(){char*p=t;while(1){int
c=getchar();if(EOF==c)break;
putchar(c^*p);if(!*++p)p=t;}}
