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
	int iteration1 = 0, iteration2 = 0;
	while (one.word[0] != '\0') {
		while (two.word[0] != '\0') {
			two.word = remove_first_char(two);
			++iteration2;
			struct morpheme_t lm = find_longest_match(one, two);
			lm.words_count = 0;
			lm.words = malloc(0);
			if (lm.morpheme[0] != '\0' && strlen(lm.morpheme) >= 2) {
				if (morpheme_list.count > 0) {
					struct word_t current, previous;
					setword(&current, lm.morpheme);
					setword(&previous, morpheme_list.list[morpheme_list.count-1].morpheme);
					if (strcmp(find_longest_match(current, previous).morpheme, lm.morpheme) != 0) {
						add_word(&lm, one_orig);
						add_word(&lm, two_orig);
						add_morpheme(&morpheme_list, lm);
						gen_regex(word1, word2, iteration1, iteration2, &lm);
					}
				} else {
					add_morpheme(&morpheme_list, lm);
					gen_regex(word1, word2, iteration1, iteration2, &lm);
				}
			} else {
				free(lm.words);
			}
		}
		free(two.word);
		two.word = strdup(word2);
		one.word = remove_first_char(one);
		++iteration1;
		iteration2 = 0;
	}
	one.word = strdup(word1);
	two.word = strdup(word2);
	one_orig.word = strdup(word1);
	two_orig.word = strdup(word2);

	free(word1);
	free(word2);

	return morpheme_list;
}
