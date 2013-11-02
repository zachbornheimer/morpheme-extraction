/*
 * ngram_t.c
 * Purpose: Generate the ngram_t data type functions
 * Usage: struct ngram_t = ngram(target, &before, &after);
 */

#ifndef NGRAM_DS
#define NGRAM_DS

#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#include "constants.h"
#include "structs.h"
#include "ngram_structs.h"

int add_ngram_element(struct word_t_array **at, int index)
{
	at[index]->elems = realloc(at[index]->elems, sizeof(struct word_t) * (++(at[index]->count) + 1));
	if (at[index]->elems == NULL)
		exit(E_REALLOC);
	return at[index]->count; 
}

struct ngram_t new_ngram(void)
{
	struct ngram_t ng;
	
	int index;
	
	for (index = 0; index < ((NGRAM_SIZE/2) + 1); ++index) {
		ng.before.at[index] = malloc(sizeof(struct word_t_array));
		ng.after.at[index]  = malloc(sizeof(struct word_t_array));

		ng.before.at[index]->count = -1;
		ng.after.at[index]->count  = -1;

		ng.before.at[index]->elems = malloc(sizeof(struct word_t) * ++(ng.before.at[index]->count));
		ng.after.at[index]->elems  = malloc(sizeof(struct word_t) * ++(ng.after.at[index]->count));
	}

	return ng;
}

void free_ngram(struct ngram_t *ng)
{
	int index;
	for (index = ((NGRAM_SIZE/2)); index >= 0; --index) {
		free(ng->after.at[index]->elems);
		free(ng->before.at[index]->elems);
		free(ng->after.at[index]);
		free(ng->before.at[index]);
	}
}

/*
 * Example Implementation:
 *
 * int main()
 *  {
 * 	char* word_elem = "word";
 * 	int index = 0;
 * 	
 * 	struct ngram_t ng;
 * 	ng = new_ngram();
 * 	setword(&ng.word, word_elem);
 * 
 * 	int elem_id = add_ngram_element(&(ng.before.at[index]), index);
 * 
 * 	setword(&ng.before.at[index]->elems[elem_id], "context");
 * 	++ng.before.at[index]->elems[elem_id].freq;
 * 
 * 	printf("%s", ng.before.at[index]->elems[elem_id].word);
 * 	free_ngram(&ng);
 * 	
 * 	return 0;
 * }
 *
 */

#endif
