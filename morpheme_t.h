/*
 * morpheme_t.h
 * Defines prototypes for morpheme_t.c
 */

#ifndef MORPHEME_T_PROTO
#define MORPHEME_T_PROTO

#include "structs.h"

struct morpheme_t find_longest_match(struct word_t, struct word_t);
void free_morpheme(struct morpheme_t*);

#endif
