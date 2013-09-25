/*
 * File: words.c
 * Author: Z. Bornheimer
 * Purpose: Define Data Structures for words
 *
 *
 * usage: words w = word("word");
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>


typedef struct words_t {
    unsigned int i;
    char *word;
} words;

/* allows for setting of the word */
void setword(words *word, char *w)
{
    word->i = strlen(w) + 1; /* include \0 */
    word->word = malloc(sizeof(words) * ++word->i);
    word->word = w;
    if (word->word == NULL) {
        printf("malloc failed.\n");
        exit(1);
    }

}

/* constructs a new word */
words word(char *w)
{
    words a = {.i = 0};
    setword(&a, w);
    return a;
}
