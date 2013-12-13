/*
 * nlp.c
 * Written by Z. Bornheimer
 * Purpose: Fulfill the Spec I wrote to fully outline a language
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

/* orderd by visual aesthetic :) */
#include "morpheme_list_t.h"
#include "morpheme_t.h"
#include "constants.h"
#include "directory.h"
#include "functions.h"
#include "alphabet.h"
#include "structs.h"
#include "ngram_t.h"
#include "word_t.h"
#include "file.h"
#include "nlp.h"

int verbose_mode;
int process_sequentially;
int process_full;
char *output_filename;
char *corpus_dir;

/* gets/sets command line arguments */
int main(const int argc, char *argv[])
{
	int i;
	verbose_mode = OFF;
	process_sequentially = OFF;
	process_full = OFF;
	output_filename = ZEDRAM_OUTPUT;
	corpus_dir = DEFAULT_PATH;
	for (i = 0; i < argc; ++i)
		if (!strcmp(argv[i], "--verbose")) {
			verbose_mode = ON;
		} else if (!strcmp(argv[i], "--process-sequentially")) {
			process_sequentially = ON;
		} else if (!strcmp(argv[i], "--serial")) {
			process_sequentially = ON;
		} else if (!strcmp(argv[i], "--sequential")) {
			process_sequentially = ON;
		} else if (!strcmp(argv[i], "--process")) {
			process_full = ON;
		} else if (!strcmp(argv[i], "--output-file")) {
			if (i+1 < argc)
				output_filename = argv[i+1];
		} else if (!strcmp(argv[i], "--corpus-dir")) {
			if (i+1 < argc)
				corpus_dir = argv[i+1];
		}

	return nlp();
}	

/* runs the nlp proc */
int nlp(void)
{
	char *f, *wd;
	int index = 0, ngram_length = 0, lex_count = 0;
	char *header;
	struct ngram_t **ng;
	struct lexical_categories_t *morphemes;

	while((f = getfiles(&index, &header))) {
		if (f != NULL && strlen(f) > 2) {
			V_PRINT("Looking for Word Delimiter...");
			wd = find_word_delimiter(&f);
			V_PRINT("Found Word Delimiter");
			if (wd != NULL && wd != 0 && errno != E_OVERRULED) {
				V_PRINT("Building NGRAMS");
				build_ngram_relationships(wd, f, &ngram_length, &ng);
				if (errno == E_OVERRULED)
					continue;
				if (process_full == ON || process_sequentially == ON) {
					V_PRINT("\nPROCESSING SEQUENTIALLY!");
					V_PRINT("Extracting Morphemes and Building Lexical Categories.");
					lex_count = find_morphemes(ng, ngram_length, header, &morphemes);
				}
			}
			free(wd);
		}
		free(header);
	}

	if (process_full == ON || process_sequentially == OFF) {
		header = "All Processed Files";
		V_PRINT("PROCESSING SIMULATNEOUSLY!");
		V_PRINT("Extracting Morphemes and Building Lexical Categories.");
		lex_count = find_morphemes(ng, ngram_length, header, &morphemes);
	}

	for (lex_count -= 1; lex_count >= 0; --lex_count) {
		struct lexical_categories_t grapheme = morphemes[lex_count];
		free(grapheme.morpheme.morpheme);
		free(grapheme.morpheme.regex);
	}

	for (ngram_length -= 1; ngram_length >= 0; --ngram_length)
		free_ngram(ng[ngram_length]);


	return 0;
}

void build_ngram_relationships(const char *wd, char *f, int *ngram_length, struct ngram_t ***ngram_full)
{
	char **arr;
	int word_count = explode_sansnull_str(&arr, f, &wd);

	int i = 0, q = 0, j = 0;
	int ngram_count = word_count;
	int make_new = 1;
	int add_loc = *ngram_length;

	struct ngram_t **ng = *ngram_full;

	if (add_loc == 0)
		ng = malloc(sizeof(struct ngram_t) * (word_count+1));
	V_PRINT("Setting NGRAM elements");
	for (i = 0; i <= word_count; ++i) {
		struct ngram_t ngram;
		if (arr[i] == NULL || strcmp(arr[i], "") == 0 || strlen(arr[i]) == 0)
			continue;

		if (i != 0) {
			for (q = 0; q < add_loc; ++q)
				if (strcmp(ng[q]->word.word, arr[i]) == 0) {
					ngram = *ng[q];
					--ngram_count;
					make_new = 0;
					break;
				} else {
					make_new = 1;
				}
		}
		if (make_new == 1) {
			ngram = new_ngram();
			if (strlen(arr[i]) > 0)
				setword(&ngram.word, arr[i]);
		}

		if (verbose_mode == ON)
			printf("  Processing Word: %d:%d     \r", i, word_count);
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
		if (make_new == 1) {
			ng[add_loc] = malloc(sizeof(struct ngram_t));
			*ng[add_loc] = ngram;
			++add_loc;
		} else {
			ng[q] = malloc(sizeof(struct ngram_t));
			*ng[q] = ngram;
		}
	}
	i = add_loc-1;
	*ngram_length = add_loc;

	int k;
	for (k = 0; k < *ngram_length; ++k) {
		struct word_t w = ng[k]->word;
		for (j = k; j < *ngram_length; ++j) {
			if (verbose_mode == ON)
				printf("Processing Similar Ngrams: Target: %d/%d| Similar: %d/%d     \r     ", k, *ngram_length, j, *ngram_length);
			if (j - k < NGRAM_SIZE)
				continue;
			struct word_t a = ng[j]->word;
			struct ngram_t one = *ng[k];
			struct ngram_t two = *ng[j];
			if (a.word != NULL && w.word != NULL && strcmp(a.word, w.word) != 0 && ngrams_similar(one, two))
				add_similar_ngram_ref(&ng[k], &ng[j]);
		}
	}

	V_PRINT("");
	*ngram_full = ng;
}

int find_morphemes(struct ngram_t **ng, int ngram_length, char *header, struct lexical_categories_t **lex_map)
{
	struct morpheme_list_t internal = {.count = 0};
	internal.list = malloc(0);

	int i, j;
	for (i = 0; i < ngram_length; ++i) {
		struct ngram_t ngram = *(ng[i]);
		for (j = 0; j < ngram.refs_count; ++j) {
			if (verbose_mode == ON)
				printf(" Processing Ngram %d/%d:%d/%d           \r", i,ngram_length-1, j, ngram.refs_count-1);
			struct ngram_t target = *ngram.refs[j];
			char *w1 = strdup(ngram.word.word);
			char *w2 = strdup(target.word.word);

			struct morpheme_t forward = find_longest_match(ngram.word, target.word);
			forward.words_count = 0;
			forward.words = malloc(0);
			gen_regex(ngram.word.word, target.word.word, 0, 0, &forward);
			add_word(&forward, ngram.word);
			add_word(&forward, target.word);

			add_morpheme(&internal, forward);
			find_internal_morphemes(ngram.word, target.word, &internal);

			ngram.word.word = w1;
			target.word.word = w2;
		}
	}
	V_PRINT("");
	struct morpheme_list_t morpheme_list = fuse_regex(internal);
	i = 0;
	struct lexical_categories_t *lex = malloc(sizeof(struct lexical_categories_t) * morpheme_list.count);
	identify_true_morphemes(&morpheme_list, &lex);
	write_to_file(output_filename, lex, morpheme_list.count, header);

	for (i = 0; i < morpheme_list.count; ++i)
		free(morpheme_list.list[i].words);
	free(morpheme_list.list);

	*lex_map = lex;

	return morpheme_list.count;
}
