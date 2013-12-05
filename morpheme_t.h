/*
 * morpheme_t.h
 * Defines prototypes for morpheme_t.c
 */

#ifndef MORPHEME_T_PROTO
#define MORPHEME_T_PROTO

#include "structs.h"

void add_word(struct morpheme_t*, struct word_t);
struct morpheme_t find_longest_match(struct word_t, struct word_t);
void free_morpheme(struct morpheme_t*);
void gen_regex(char *, char *, int, int, struct morpheme_t*);

#endif
