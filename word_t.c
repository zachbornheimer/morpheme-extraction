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
#include "functions.h"
#include "word_t.h"

/* allows for setting of the word */
void setword(struct word_t *word, char *w)
{
	word->i = strlen(w);
	word->freq = 1; /* initialized */
	word->word = strdup(w);
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


char* remove_first_char(struct word_t w)
{
	char *word = strdup(w.word);
	char *finalized = reverse(word);
	finalized[strlen(finalized)-1] = '\0';
	return reverse(finalized);
}

#endif
