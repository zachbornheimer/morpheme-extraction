/*
 * morpheme_list_t.c
 * Defines the larger morpheme_list struct
 */

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "structs.h"
#include "morpheme_list_t.h"
#include "morpheme_t.h"
#include "word_t.h"

void add_morpheme(struct morpheme_list_t *morphemes, struct morpheme_t morpheme)
{
	int size = sizeof(struct morpheme_t) * ++(morphemes->count);
	morphemes->list = realloc(morphemes->list, size);
	morphemes->list[morphemes->count-1] = morpheme;
}

struct morpheme_list_t find_internal_morphemes(struct word_t one_orig, struct word_t two_orig)
{
	char *word1 = strdup(one_orig.word), *word2 = strdup(two_orig.word);
	struct word_t one = one_orig, two = two_orig;
	struct morpheme_list_t morpheme_list = {.count = 0};
	morpheme_list.list = malloc(0);
	while (one.word[0] != '\0') {
		while (two.word[0] != '\0') {
			two.word = remove_first_char(two);
			struct morpheme_t lm = find_longest_match(one, two);
			if (lm.morpheme[0] != '\0' && strlen(lm.morpheme) >= 2)
				add_morpheme(&morpheme_list, lm);
		}
		two.word = strdup(word2);
		one.word = remove_first_char(one);
	}
	one.word = strdup(word1);
	two.word = strdup(word2);
	one_orig.word = strdup(word1);
	two_orig.word = strdup(word2);

	return morpheme_list;
}
