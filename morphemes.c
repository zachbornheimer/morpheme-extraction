/*
 * morphemes.c
 *
 * Morpheme Extraction Algorithm
 * by Z. Bornheimer
 *
 * 
 */

#include <stdio.h>
#include "prototypes.h"
#include "words_t.c"

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

    struct words_t *w;
    int i;
    for (i = 0; i < 100; ++i) {
        w = malloc(sizeof(struct words_t) * (i+1));
        w[i] = word("test");
    }
    int f = i -1;
        printf("%s%d",w[f].word,f);
}
