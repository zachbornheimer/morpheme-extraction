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
	int freq;
};

struct lexical_categories_t {
	struct morpheme_t morpheme;
	int ref_count;
	struct ngram_t **refs;
};

/* end lexical_categories_t def. */

#endif
