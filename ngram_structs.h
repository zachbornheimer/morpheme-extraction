/*
 * ngram_structs.h
 * Unify N-gram-based Struct Definitions in one file
 */

#ifndef NGRAM_STRUCT_DEFS
#define NGRAM_STRUCT_DEFS

#include "constants.h"

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
};


#endif
