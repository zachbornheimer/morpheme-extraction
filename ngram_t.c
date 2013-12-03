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
#include "functions.h"

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
	
	for (index = 0; index < (NGRAM_SIZE/2); ++index) {
		int wta_size = sizeof(struct word_t_array);
		ng.before.at[index] = malloc(wta_size);
		ng.after.at[index]  = malloc(wta_size);

		ng.before.at[index]->count = 0;
		ng.after.at[index]->count  = 0;

		ng.before.at[index]->elems = malloc(sizeof(struct word_t));
		ng.after.at[index]->elems  = malloc(sizeof(struct word_t));
	}

	ng.refs_count = 0;

	return ng;
}

void free_ngram(struct ngram_t *ng)
{
	int index;
	for (index = ((NGRAM_SIZE/2) - 1); index >= 0; --index) {
		/*free(ng->after.at[index]->elems);
		free(ng->before.at[index]->elems);
		free(ng->after.at[index]);
		free(ng->before.at[index]);*/
	}
	free(ng);
}



void quicksort_ngram_array(struct ngram_t ***ng_real, int ngram_size, int type)
{
	struct ngram_t **ng = *ng_real;
	int i = 0, k = 0;
	for (k = 0; k < ngram_size; ++k) {
		int condition, mark = k;
		for (i = 0; i < mark; ++i) {
			if (type == ASCENDING)
				condition = (ng[i]->word.freq > ng[mark]->word.freq);
			else
				condition = (ng[i]->word.freq < ng[mark]->word.freq);
				if (condition) {
					struct ngram_t marked = (*ng)[mark];
					(*ng)[mark] = (*ng)[i];
					(*ng)[i] = marked;
				}
			}
		for (i = ngram_size-1; i > mark; --i) {
			if (type == ASCENDING)
				condition = (ng[i]->word.freq < ng[mark]->word.freq);
			else
				condition = (ng[i]->word.freq > ng[mark]->word.freq);
			if (condition) {
				struct ngram_t marked = (*ng)[mark];
				(*ng)[mark] = (*ng)[i];
				(*ng)[i] = marked;
			}
		}
	}
	*ng_real = ng;
}

void uniqify(struct ngram_half_array *bora)
{
	/* bora = before or after */
	int i, j, k;
	for (i = 0; i < (NGRAM_SIZE/2); ++i) {
		int count = bora->at[i]->count;
		for (j = 0; j < count; ++j) {
			struct word_t *elems = malloc(sizeof(struct word_t) * (bora->at[i]->count + 1));
			int elems_count = 0;
			elems[elems_count++] = bora->at[i]->elems[j];
			struct word_t elem = elems[elems_count-1];
			for (k = j+1; k < count - 1; ++k) {
				struct word_t comp = bora->at[i]->elems[k];
				if (comp.freq != 1)
					continue;
				if (strcmp(elem.word, comp.word) == 0)
					elem.freq += bora->at[i]->elems[k].freq;
				else
					elems[elems_count++] = bora->at[i]->elems[k];
			}
			bora->at[i]->elems = elems;
		}
	}
}

/*
void quicksort_ngram(struct ngram_t **ng_real, int type)
{
	struct ngram_t **ng = *ng_real;
	int i = 0, k = 0;
	// before
	for (k = 0; k < NGRAM_SIZE/2; ++k) {
		int condition, mark = k;
		for (i = 0; i < mark; ++i) {
			if (type == ASCENDING)
				condition = (ng[i]->word.freq > ng[mark]->word.freq);
			else
				condition = (ng[i]->word.freq < ng[mark]->word.freq);
				if (condition) {
					struct ngram_t marked = (*ng)[mark];
					(*ng)[mark] = (*ng)[i];
					(*ng)[i] = marked;
				}
			}
		// after
		for (i = ngram_size-1; i > mark; --i) {
			if (type == ASCENDING)
				condition = (ng[i]->word.freq < ng[mark]->word.freq);
			else
				condition = (ng[i]->word.freq > ng[mark]->word.freq);
			if (condition) {
				struct ngram_t marked = (*ng)[mark];
				(*ng)[mark] = (*ng)[i];
				(*ng)[i] = marked;
			}
		}
	}
	*ng_real = ng;
}
*/

int ngrams_similar(struct ngram_t a, struct ngram_t b)
{

	int percent_similar = 0;
	int elems_count = 0;
	int arr_count = 0;
	int count = 0;

	int i, j;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		elems_count += a.before.at[i]->count;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		elems_count += a.after.at[i]->count;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		elems_count += b.before.at[i]->count;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		elems_count += b.after.at[i]->count;

	char **arr = malloc(sizeof(char*) * elems_count);

	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		for (j = 0; j < a.before.at[i]->count; ++j)
			if (strcmp(a.before.at[i]->elems[j].word, "") != 0)
				arr[arr_count++] = a.before.at[i]->elems[j].word;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		for (j = 0; j < a.after.at[i]->count; ++j)
			if (strcmp(a.after.at[i]->elems[j].word, "") != 0)
				arr[arr_count++] = a.after.at[i]->elems[j].word;

	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		for (j = 0; j < b.before.at[i]->count; ++j)
			if (strcmp(b.before.at[i]->elems[j].word, "") != 0)
				arr[arr_count++] = b.before.at[i]->elems[j].word;
	for (i = 0; i < (NGRAM_SIZE/2); ++i)
		for (j = 0; j < b.after.at[i]->count; ++j)
			if (strcmp(b.after.at[i]->elems[j].word, "") != 0)
				arr[arr_count++] = b.after.at[i]->elems[j].word;
       	count = uniq_words(arr, elems_count);
	free(arr);

	arr_count -= 1;
	percent_similar = (double) 100.00 * (((double)arr_count - (double)count)/(double)arr_count);


	if (percent_similar >= THRESHOLD_SIMILAR_NGRAMS)
		return 1;
	else
		return 0;
}

void add_similar_ngram_ref(struct ngram_t **ngram, struct ngram_t **ngram_ref)
{
	int num_refs = (*ngram)->refs_count;
	int memsize = (sizeof(struct ngram**) * (num_refs+1));

	if (num_refs == 0) {
		(*ngram)->refs = malloc(memsize);
	} else {
		(*ngram)->refs = realloc((*ngram)->refs, memsize);
		if ((*ngram)->refs == NULL)
			exit(E_REALLOC);
	}

	(*ngram)->refs[(*ngram)->refs_count] = *ngram_ref;
	++(*ngram)->refs_count;

}


#endif
