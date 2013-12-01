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
#include "directory.h"
#include "functions.h"
#include "alphabet.h"
#include "structs.h"
#include "ngram_t.h"
#include "word_t.h"
#include "file.h"

int verbose_mode;

struct lexical_categories_t* find_morphemes(struct ngram_t**, int);
struct ngram_t** build_ngram_relationships(char*, char*, int*);
int nlp(void);

int main(int argc, char *argv[])
{
	int i;
	verbose_mode = OFF;
	for (i = 0; i < argc; ++i) {
		if (!strcmp(argv[i], "--verbose"))
			verbose_mode = ON;
	}

	return nlp();
}	

int nlp(void)
{
	char *f;
	int index = 0;
	char *wd;
	int ngram_length = 0;
	struct ngram_t **ng;
	struct lexical_categories_t *morphemes;
	/*struct ngram_t **ng_to_merge;*/
	while((f = getfiles(&index))) {
		if (f != NULL && strlen(f) > 2) {
			V_PRINT("Looking for Word Delimiter...");
			wd = find_word_delimiter(&f);
			V_PRINT("Found Word Delimiter");
			if (wd != NULL && wd != 0 && errno != E_OVERRULED) {
				V_PRINT("Building NGRAMS");
				ng = build_ngram_relationships(wd, f, &ngram_length);
				if (errno == E_OVERRULED)
					continue;
				V_PRINT("Extracting Morphemes and Building Lexical Categories.");
				morphemes = find_morphemes(ng, ngram_length);
				/*V_PRINT("Appending Current File n-gram with previous");
				ngram_append(&ng_to_merge, &ng);*/
			}
		}
	}

	for (; ngram_length >= 0; --ngram_length)
		free_ngram(ng[ngram_length]);


	return 0;
}

struct ngram_t** build_ngram_relationships(char *wd, char *f, int *ngram_length)
{
	char **arr;
	int word_count = explode_sansnull_str(&arr, f, &wd);
	
	int i, q, j;
	int hash_elem = -1;
	int hash_index = -1;
	int ngram_count = word_count;
	int make_new = 1;

	struct ngram_t **ng;
	struct word_t *w_freq = malloc(sizeof(struct word_t) * (word_count+1));
	
	ng = malloc(sizeof(struct ngram_t) * (word_count+1));
	V_PRINT("Setting NGRAM elements");
	for (i = 0; i <= word_count; ++i) {
		struct ngram_t ngram;
		if (arr[i] == NULL || strcmp(arr[i], "") == 0 || strlen(arr[i]) == 0)
			continue;
		if (i != 0) {
			for (q = 0; q < i; ++q) {
				if (ng[q]->word.word != NULL && strcmp(ng[q]->word.word, arr[i]) == 0) {
					ngram = *ng[q];
					--ngram_count;
					make_new = 0;
					break;
				}
			}
		}
		if (make_new == 1) {
			ngram = new_ngram();
			if (strlen(arr[i]) > 0)
				setword(&ngram.word, arr[i]);
		}

		if (verbose_mode == ON)
			printf("  Processing Word: %d:%d\r", i, word_count);
		if (strlen(arr[i]) > 0) {
			++ngram.word.freq;

			for (j = 0; j < (NGRAM_SIZE/2); ++j) {
				int index = i - (NGRAM_SIZE/2) + j;
				int elem_id;

				if (index > 0 && strlen(arr[index])>0) {
					elem_id	= add_ngram_element(&(ngram.before.at[j]), j);
					setword(&(ngram.before.at[j]->elems[elem_id]), arr[(index)]);
				}

				if ((index+(NGRAM_SIZE/2)+1)<=word_count && strlen(arr[index+(NGRAM_SIZE/2)+1])>0) {
					elem_id = add_ngram_element(&(ngram.after.at[j]), j);
					setword(&(ngram.after.at[j]->elems[elem_id]), arr[index+(NGRAM_SIZE/2)+1]);
				}

			}

		}
		ng[i] = malloc(sizeof(struct ngram_t));
		*ng[i] = ngram;
	}
	*ngram_length += ngram_count;

/*	V_PRINT("Uniqifying each ngram");

	for (q = 0; q <= word_count; q++) {
		if (verbose_mode == ON)
			printf("  Uniqifying ngram: %d/%d\r", q, word_count);
		if (q < (NGRAM_SIZE/2)) {
			uniqify(&ng[q]->before);
		}
		if (word_count - q > NGRAM_SIZE)
			uniqify(&ng[q]->after);
	}
*/


	int k;
	V_PRINT("Processing Similar Ngrams...");
	for (k = 0; k <= *ngram_length; ++k) {
		struct word_t w = ng[k]->word;
		for (j = 0; j <= *ngram_length; ++j) {
			if (verbose_mode == ON)
				printf("Processing Similar Ngrams: Target: %d/%d| Similar: %d/%d\r", k, *ngram_length, j, *ngram_length);
			if (j - k < NGRAM_SIZE)
				continue;
			struct word_t a = ng[j]->word;
			struct ngram_t one = *ng[k];
			struct ngram_t two = *ng[j];
			if (a.word != NULL && w.word != NULL && strcmp(a.word, w.word) != 0 && ngrams_similar(one, two))
				add_similar_ngram_ref(&ng[k], &ng[j]);
		}
	}

	return ng;

}

struct lexical_categories_t* find_morphemes(struct ngram_t **ng, int ngram_length)
{
	struct lexical_categories_t *lex = malloc(sizeof(struct lexical_categories_t));
	struct morpheme_t *morphemes;
	int elem_count = 0;

	int i, j;
	for (i = 0; i <= ngram_length; ++i) {
		struct ngram_t ngram = *ng[i];
		for (j = 0; j < ngram.refs_count; ++j) {
		/*	struct ngram_t target = *ngram.refs[j];
			struct morpheme_t forward = find_longest_match(ngram->word, target->word);
			struct morpheme_t backward = find_longest_match(reverse_word(ngram->word), reverse_word(target->word));
			backward.regex[0] = '$';
			backward.regex = reverse(backward.regex);
			struct morpheme_t *internal = find_internal_morphemes(ngram->word, target->word);
			*/
			//merge_morpheme_list(&morphemes, forward, backward, internal, &elem_count);
		}
	}
/*	for (int j = 0; j < elem_count; ++j) {
		regexify(&morphemes[j]);
	}
	uniq_morpheme_t(&morphemes);

	identify_true_morphemes(&morphemes, &lex)
		store_morphemes(&lex);
*/


	return lex;
}
