/*
 * structs.h
 * Unify Struct Definitions in one file
 */

#ifndef STRUCT_DEFS
#define STRUCT_DEFS

#include "constants.h"

struct char_doubleton {
	char c;
	int freq;
};

struct word_t {
	unsigned int i;
	int freq;
	char *word;
};

/* ngram structs */

struct word_t_array {
	struct word_t *elems;
	unsigned int count;
};

struct ngram_half_array {
	struct word_t_array *at[(NGRAM_SIZE/2)+1];
};

struct ngram_t {
	struct word_t word;
	struct ngram_half_array before;
	struct ngram_half_array after;
	int refs_count;
	struct ngram_t **refs; 
};

/* end ngram structs */

/* lexical_categories_t struct */

struct morpheme_t {
	char *regex;
	char *morpheme;
	int freq;
	int words_count;
	struct word_t *words;
	int type;
	char *front_regex;
	char **front_regex_arr;
	char *back_regex;
	char **back_regex_arr;
	int front_regex_arr_index;
	int back_regex_arr_index;
};

struct morpheme_list_t {
	struct morpheme_t *list;
	int count;
};

struct lexical_categories_t {
	struct morpheme_t morpheme;
	int ref_count;
	struct ngram_t **refs;
};

/* end lexical_categories_t def. */

#endif
