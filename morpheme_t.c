/* * morpheme_t.c
 * The functions related to extracting morphemes
 */

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "structs.h"
#include "morpheme_t.h"
#include "morpheme_list_t.h"

struct morpheme_t find_longest_match(struct word_t one, struct word_t two)
{
	struct word_t longest, shortest;
	struct morpheme_t empty = {.morpheme = "\0", .freq = 1};
	int real_length = 0;

	if (one.i >= two.i) {
		longest = one;
		shortest = two;
	} else {
		longest = two;
		shortest = one;
	}

	char *str = malloc(sizeof(char) * (shortest.i + 2));
	while (shortest.word[real_length] != '\0' && shortest.word[real_length] == longest.word[real_length]) {
		str[real_length] = shortest.word[real_length];
		++real_length;
	}
	str[real_length] = '\0';
	
	/*if (real_length == 0)
		printf("ZERO!, w1: %s, w2: %s\n", shortest.word, longest.word);*/

	if (real_length == 0)
		return empty;



	str = realloc(str, (sizeof(char) * real_length+1));
	if (str == NULL)
		exit(E_REALLOC);

	struct morpheme_t longest_match;

	longest_match.freq = 1;
	longest_match.morpheme = str;
	
	return longest_match;

}

void free_morpheme(struct morpheme_t *morpheme)
{
	free(morpheme->regex);
	free(morpheme);
}
