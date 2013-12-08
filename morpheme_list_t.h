/*
 * morpheme_list_t.h
 * Defines the prototypes for morpheme_list.c
 */

#ifndef MORPHEME_LIST_PROTO
#define MORPHEME_LIST_PROTO

#include "structs.h"

void add_morpheme(struct morpheme_list_t *morphemes, struct morpheme_t morpheme);
void find_internal_morphemes(struct word_t, struct word_t, struct morpheme_list_t*);
struct morpheme_list_t fuse_regex(struct morpheme_list_t);

#endif
