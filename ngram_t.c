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

/*
#include "word_t.c"
*/

int add_ngram_element(struct word_t_array **at_index, int index)
{
	unsigned int size = sizeof(struct word_t) * (++((*at_index)->count) + 1);
	(*at_index)->elems = realloc((*at_index)->elems, size);
	if ((*at_index)->elems == NULL)
		exit(E_REALLOC);
	return (*at_index)->count - 1; 
}

struct ngram_t new_ngram(void)
{
	struct ngram_t ng;
	
	int index;
	
	for (index = 0; index < ((NGRAM_SIZE/2)); ++index) {
		int wta_size = sizeof(struct word_t_array);
		ng.before.at[index] = malloc(wta_size);
		ng.after.at[index]  = malloc(wta_size);

		ng.before.at[index]->count = 0;
		ng.after.at[index]->count  = 0;

		ng.before.at[index]->elems = malloc(sizeof(struct word_t));
		ng.after.at[index]->elems  = malloc(sizeof(struct word_t));
	}

	return ng;
}

void free_ngram(struct ngram_t *ng)
{
	int index;
	for (index = ((NGRAM_SIZE/2) - 1); index >= 0; --index) {
		//free(ng->after.at[index]->elems);
		//free(ng->before.at[index]->elems);
		//free(ng->after.at[index]);
		//free(ng->before.at[index]);
	}
	free(ng);
}

/*

 int main()
  {
 	char* word_elem = "word";
 	int index = 0;
 	
 	struct ngram_t ng;
 	ng = new_ngram();
 	setword(&ng.word, word_elem);
 
 	int elem_id = add_ngram_element(&(ng.before.at[index]), index);
 
 	setword(&ng.before.at[index]->elems[elem_id], "context");
 	++ng.before.at[index]->elems[elem_id].freq;
 
 	printf("%s", ng.before.at[index]->elems[elem_id].word);
 	free_ngram(&ng);
 	
 	return 0;
 }

*/

#endif
