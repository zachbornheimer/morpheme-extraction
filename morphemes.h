/*
 * morphemes.h
 *
 * Morpheme Extraction Algorithm
 * by Z. Bornheimer
 *
 * 
 */

#include <stdio.h>
#include "prototypes.h"
#include "words.h"

void morpheme_extraction()
{
    /* 
     Pseudo-code:

        input should be something like a
            words *w;  needs to dynamically malloc
     */

}

int main()
{

    words *w;
    int i;
    for (i = 0; i < 100; ++i) {
        w = malloc(sizeof(words) * (i+1));
        w[i] = word("test");
    }
    int f = i -1;
        printf("%s%d",w[f].word,f);
}
