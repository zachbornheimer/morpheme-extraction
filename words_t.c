/*
 * File: words_t.c
 * Purpose: Define Data Structures for words
 * Usage: words w = word("word");
 */

#ifndef WORD_DS
#define WORDS_DS

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "structs.h"

/* allows for setting of the word */
void setword(struct words_t *word, char *w)
{
    word->i = strlen(w); /* include \0 */
    word->word = malloc(sizeof(struct words_t) * ++word->i);
    word->word = w;
    if (word->word == NULL) {
        printf("malloc failed.\n");
        exit(1);
    }

}

/* constructs a new word */
struct words_t word(char *w)
{
    struct words_t a = {.i = 0};
    setword(&a, w);
    return a;
}

#endif
