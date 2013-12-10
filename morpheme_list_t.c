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
#include "functions.h"

void add_morpheme(struct morpheme_list_t *morphemes, struct morpheme_t morpheme)
{
	int size = sizeof(struct morpheme_t) * ++(morphemes->count);
	morphemes->list = realloc(morphemes->list, size);
	morphemes->list[morphemes->count-1] = morpheme;
}

void find_internal_morphemes(struct word_t one_orig, struct word_t two_orig, struct morpheme_list_t *internal)
{
	char *word1 = strdup(one_orig.word), *word2 = strdup(two_orig.word);
	struct word_t one = one_orig, two = two_orig;
	int iteration1 = 0, iteration2 = 0;
	int lookback = 1;
	while (one.word[0] != '\0') {
		while (two.word[0] != '\0') {
			two.word = remove_first_char(two);
			++iteration2;
			if (strlen(two.word) < 2)
				break;
			struct morpheme_t lm = find_longest_match(one, two);
			lm.words_count = 0;
			lm.words = malloc(0);
			if (lm.morpheme[0] != '\0' && strlen(lm.morpheme) >= 2) {
				if ((*internal).count > 0) {
					struct word_t current, previous;
					setword(&current, lm.morpheme);
					setword(&previous, (*internal).list[(*internal).count-1].morpheme);
					char *test = strdup(previous.word);
					test = reverse(test);
					if (((int)strlen(test)-lookback) >= 1)
						if (test[strlen(test)-lookback] != 0)
							test[strlen(test)-lookback] = '\0';
					test = reverse(test);
					if (strcmp(find_longest_match(current, previous).morpheme, lm.morpheme) != 0) {
						if (strcmp(test, current.word) != 0) {
							add_word(&lm, one_orig);
							add_word(&lm, two_orig);
							gen_regex(word1, word2, iteration1, iteration2, &lm);
							if (lm.regex != NULL)
								add_morpheme(internal, lm);
							lookback = 1;
						} else {
							++lookback;
							free(lm.words);
						}
					}
				} else {
					gen_regex(word1, word2, iteration1, iteration2, &lm);
					if (lm.regex != NULL)
						add_morpheme(internal, lm);
					lookback = 1;
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
}

struct morpheme_list_t fuse_regex(struct morpheme_list_t original)
{
	/* Proc:
	 *     In Char Array, Morpheme? If not, add the morpheme as is.
	 *         If yes, fuse the regex with the existing morpheme
	 *
	 *     Regex Fusion:
	 *         ([fw][ro][or][nd][t])morpheme([bw][ao][cr][kd])
	 *         + ([b][l][o][c][k])morpheme([t][e][s][t])
	 *
	 *         ([fw][ro][or][nd][t]) + ([b][l][o][c][k])
	 *         Create new, if block, while (!= ']') add char
	 *         when ']', proc block in 2nd list, while != ']'.
	 *         when ']' for both, add ']' and move on.
	 *         Repeat until ')'.
	 *
	 *         Do this for the back end as well.
	 */

	int i;
	char **morpheme_list = malloc(sizeof(char*) * original.count);
	int *index_lookup = malloc(sizeof(int) * original.count);
	struct morpheme_list_t new = {.count = 0};
	new.list = malloc(sizeof(struct morpheme_t) * original.count);
	int morpheme_list_count = 0;
	int current_index = -1;
	char **regex_front;
	char **regex_back;
	for (i = 0; i < original.count; ++i) {
		if (strcmp(original.list[i].morpheme, "") == 0)
			continue;
		if (morpheme_list_count == 0) {
			morpheme_list[morpheme_list_count] = original.list[i].morpheme;
			add_morpheme(&new, original.list[i]);
			index_lookup[morpheme_list_count] = i;
			++morpheme_list_count;
			continue;
		}
		char *morpheme = original.list[i].morpheme;
		if ((current_index = in_char_array(morpheme, morpheme_list, morpheme_list_count)) != -1) {
			struct morpheme_t dup = original.list[index_lookup[current_index]];
			struct morpheme_t orig = new.list[current_index];

			if (orig.type == STEM)
				continue;

			struct morpheme_t working = new.list[current_index];
			/* Step One, Extract Parenthetical Sections */
			char *dup_regex = dup.regex;
			char *orig_regex = orig.regex;
			int orig_front_index = orig.front_regex_arr_index;
			int dup_front_index = dup.front_regex_arr_index;
			int orig_back_index = orig.back_regex_arr_index;
			int dup_back_index = dup.back_regex_arr_index;

			if (strcmp(orig_regex, dup_regex) == 0) {
				++new.list[current_index].freq;
				continue;
			}

			regex_front = malloc(sizeof(char*) * ((dup_front_index + 1) + (orig_front_index + 1) + 1));
			regex_back = malloc(sizeof(char*) * ((dup_back_index + 1) + (orig_back_index + 1) + 1));

			/* Step Two, Merge Blocks [] */

			/*
			 * Each Block is stored in an char* array.
			 * If I modify each of those arrays, I can then reconstruct.
			 */
			int j = 0;

			while(1) {
				int size = 2;
				int longest = (dup_front_index > orig_front_index? dup_front_index : orig_front_index);

				if (j <= orig_front_index)
					if (orig.front_regex_arr[j][0] > 0)
						size += (int) strlen(orig.front_regex_arr[j]);

				if (j <= dup_front_index)
					if (dup.front_regex_arr[j] != 0 && strlen(dup.front_regex_arr[j]) != 0)
						size += (int) strlen(dup.front_regex_arr[j]);

				regex_front[j] = malloc(sizeof(char) * size);
				regex_front[j][0] = '\0';

				if (size == 2)
					break;
				if (j >= longest-orig_front_index)
					if (orig_front_index != -1)
						strcat(regex_front[j], orig.front_regex_arr[j]);
				if (j >= longest-dup_front_index)
					if (dup_front_index != -1)
						strcat(regex_front[j], dup.front_regex_arr[j]);
				++j;

			}
			working.front_regex_arr_index = j-1;
			working.front_regex_arr = regex_front;

			j = 0;
			while(1) {
				int size = 2;
				int longest = (dup_back_index > orig_back_index? dup_back_index : orig_back_index);

				if (j <= orig_back_index)
					if (orig.back_regex_arr[j][0] > 0)
						size += (int) strlen(orig.back_regex_arr[j]);
				if (j <= dup_back_index)
					if (dup.back_regex_arr[j] != 0 && strlen(dup.back_regex_arr[j]) != 0)
						size += (int) strlen(dup.back_regex_arr[j]);

				regex_back[j] = malloc(sizeof(char) * size);
				regex_back[j][0] = '\0';
				if (size == 2)
					break;
				if (j >= longest-orig_back_index)
					if (orig_back_index != -1)
						strcat(regex_back[j], orig.back_regex_arr[j]);
				if (j >= longest-dup_back_index)
					if (dup_back_index != -1)
						strcat(regex_back[j], dup.back_regex_arr[j]);
				++j;
			}
			working.back_regex_arr_index = j-1;
			working.back_regex_arr = regex_back;

			/* Step Three, Reconstruct and restore */

			merge_rules(&working);
			++working.freq;
			new.list[current_index] = working;

		} else {
			morpheme_list[morpheme_list_count] = original.list[i].morpheme;
			add_morpheme(&new, original.list[i]);
			index_lookup[morpheme_list_count] = i;
			++morpheme_list_count;
		}
		//current_index = -1;
	}
	return new;
	//for (i = 0; i < morpheme_list_count; ++i)
	//	free(morpheme_list[i]);
	//	free(morpheme_list);
}

void identify_true_morphemes(struct morpheme_list_t *list, struct lexical_categories_t **lex)
{
	int i = 0;
	for (i = 0; i < list->count; ++i) {
		const struct morpheme_t morpheme = list->list[i];
		int freq_thresh = (double) list->count * (double) (THRESHOLD_CONFIRMATION/100);
		if (morpheme.freq <= freq_thresh && morpheme.type != STEM)
			continue;

		if (morpheme.type != STEM) {

			/* Prefix Identification */
			if (strlen(morpheme.back_regex) == 0 || strcmp(morpheme.back_regex, "()") == 0) {
				if (strcmp(morpheme.front_regex, "^") == 0)
					list->list[i].type = PREFIX;
			} else {
				if (strcmp(morpheme.front_regex, "()") == 0)
					list->list[i].type = PREFIX;
			}

			/* Suffix Identification */
			if (strlen(morpheme.front_regex) == 0 || strcmp(morpheme.front_regex, "()") == 0) {
				if (strcmp(morpheme.back_regex, "$") == 0)
					list->list[i].type = SUFFIX;
			} else {
				if (strcmp(morpheme.back_regex, "()") == 0)
					list->list[i].type = SUFFIX;
			}

			/* Circumfix Identification */
			int j = 0, k = 0;
			for (j = 0; j < i; ++j) {
				int combined = -1;
				if ((list->list[i].type == SUFFIX || list->list[i].type == PREFIX))
					break;
				if (!((list->list[i].type == SUFFIX && list->list[j].type == PREFIX) || (list->list[i].type == PREFIX && list->list[j].type == SUFFIX)))
					continue;
				if (list->list[j].freq == list->list[i].freq) {
					char **list_combined = malloc(sizeof(char*) * (list->list[i].words_count + list->list[j].words_count));
					for (k = 0; k < list->list[j].words_count; ++k)
						list_combined[++combined] = list->list[j].words[k].word;
					for (k = 0; k < list->list[i].words_count; ++k)
						list_combined[++combined] = list->list[i].words[k].word;

					int count = uniq_words(list_combined, combined);
					int percent_similar = (double) 100.00 * (((double)combined - (double)count)/(double)combined);
					if (percent_similar >= THRESHOLD_CIRCUMFIX) {
						list->list[i].type = CIRCUMFIX;
						list->list[j].type = CIRCUMFIX;
						if (list->list[i].type == PREFIX) {
							list->list[i].front_regex = list->list[i].regex;
							list->list[i].back_regex = list->list[j].regex;
							list->list[j].front_regex = list->list[i].regex;
							list->list[j].back_regex = list->list[j].regex;
							char *grapheme = malloc(sizeof(char) * ((strlen(list->list[i].morpheme) + strlen(list->list[j].morpheme)) + 5));
							grapheme[0] = '\0';
							strcat(grapheme, list->list[i].morpheme);
							strcat(grapheme, "(.*)");
							strcat(grapheme, list->list[j].morpheme);
							list->list[i].morpheme = grapheme;
							list->list[j].morpheme = grapheme;

						} else {
							list->list[i].front_regex = list->list[j].regex;
							list->list[i].back_regex = list->list[i].regex;
							list->list[j].front_regex = list->list[j].regex;
							list->list[j].back_regex = list->list[i].regex;
							char *grapheme = malloc(sizeof(char) * ((strlen(list->list[i].morpheme) + strlen(list->list[j].morpheme)) + 5));
							grapheme[0] = '\0';
							strcat(grapheme, list->list[j].morpheme);
							strcat(grapheme, "(.*)");
							strcat(grapheme, list->list[i].morpheme);
							list->list[i].morpheme = grapheme;
							list->list[j].morpheme = grapheme;

						}
					}
				}
			}
		}

		/* Infix Identification */
		if (list->list[i].type == UNDEF)
			list->list[i].type = INFIX;

		(*lex)[i].morpheme = list->list[i];
	}
}
