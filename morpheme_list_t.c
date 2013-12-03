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

struct morpheme_list_t find_internal_morphemes(struct word_t *one_orig, struct word_t *two_orig)
{
	struct word_t hold_one = word_t_dup(*one_orig);
	struct word_t hold_two = word_t_dup(*two_orig);
	char *word1 = strdup(one_orig->word), *word2 = strdup(two_orig->word);
	struct word_t one = hold_one, two = hold_two;
	struct morpheme_list_t morpheme_list = {.count = 0};
	morpheme_list.list = malloc(0);
	while (one.word[0] != '\0') {
		while (two.word[0] != '\0') {
			remove_first_char(&two);
			struct morpheme_t lm = find_longest_match(one, two);
			if (lm.morpheme[0] != '\0' && strlen(lm.morpheme) >= 2) {
				printf("MORPHEME: %s\n", lm.morpheme);
				add_morpheme(&morpheme_list, lm);
			}
		}
		two.word = strdup(word2);
		remove_first_char(&one);
	}
	one.word = strdup(word1);
	two.word = strdup(word2);
	one_orig->word = strdup(word1);
	two_orig->word = strdup(word2);

	printf("Backward: %s, %s\n", one_orig->word, two_orig->word);

	//printf("Morpheme: %d\n", morpheme_list.count);
	return morpheme_list;
}
