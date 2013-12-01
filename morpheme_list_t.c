/*
 * morpheme_list_t.c
 * Defines the larger morpheme_list struct
 */

#include <stdlib.h>
#include "structs.h"
#include "morpheme_list_t.h"

void add_morpheme(struct morpheme_list_t *morphemes, struct morpheme_t morpheme)
{
	morphemes->list = realloc(morphemes->list, sizeof(struct morpheme_t) * ++(morphemes->count));
	morphemes->list[morphemes->count] = morpheme;
}
