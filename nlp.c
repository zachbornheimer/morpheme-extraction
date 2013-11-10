/*
 * nlp.c
 * Written by Z. Bornheimer
 * Purpose: Fulfill the Spec I wrote to fully outline a language
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include "constants.h"
#include "prototypes.h"
#include "structs.h"
#include "ngram_structs.h"

void build_ngrams(char*, char*);

int main(void)
{
	char *f;
	int index = 0;
	char *wd;
	while((f = getfiles(&index))) {
		if (f != NULL && strlen(f) > 2) {
			wd = find_word_delimiter(&f);
			if (wd != NULL && wd != 0 && errno != E_OVERRULED) {
				build_ngrams(wd, f);
				free(f);
			}
		}
	}

	return 0;
}

void build_ngrams(char *wd, char *f)
{
	char **arr;
	int word_count = explode_sansnull_str(&arr, f, &wd);
	int i;

	struct ngram_t **ng;
	ng = malloc(sizeof(struct ngram_t) * word_count);
	for (i = 0; i <= word_count; ++i) {
		struct ngram_t ngram = new_ngram();
		if (strlen(arr[i]) > 0) {
			setword(&ngram.word, arr[i]);
			int j;
			for (j = 0; j < (NGRAM_SIZE/2); ++j) {
				int index = i - (NGRAM_SIZE/2) + j;
				if ( index >= 0 && arr[index+(NGRAM_SIZE/2)] != '\0' && arr[index] != '\0') {

					int elem_id = add_ngram_element(&(ngram.before.at[j]), j);
					setword(&(ngram.before.at[j]->elems[elem_id]), arr[index+(NGRAM_SIZE/2)]);


					elem_id = add_ngram_element(&(ngram.after.at[j]), j);
					setword(&(ngram.after.at[j]->elems[elem_id]), arr[(index)]);

				}
			}
			ng[i] = malloc(sizeof(struct ngram_t));
			*ng[i] = ngram;
		}
	}


	for (; i >= 0; --i)
		free_ngram(ng[i]);
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


}
