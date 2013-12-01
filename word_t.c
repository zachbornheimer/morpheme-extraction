/*
 * File: word_t.c
 * Purpose: Define Data Structures for words
 * Usage: struct word_t w = word("word");
 */

#ifndef WORD_DS
#define WORDS_DS

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "structs.h"

/* allows for setting of the word */
void setword(struct word_t *word, char *w)
{
	word->i = strlen(w);
	word->freq = 1; /* initialized */
	/*word->word = malloc(sizeof(char) * word->i);*/
	word->word = w;
	strcat(word->word, "\0");
	if (word->word == NULL) {
		printf("malloc failed.\n");
		exit(1);
	}

}

/* constructs a new word */
struct word_t word(char *w)
{
	struct word_t a = {.i = 0};
	setword(&a, w);
	return a;
}


struct word_t reverse_word(struct word_t w)
{
	struct word_t temp;
	temp.i = w.i;
	temp.freq = w.freq;
	temp.word = reverse(w.word);

	return temp;
}

#endif
