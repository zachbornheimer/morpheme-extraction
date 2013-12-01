/*
 * morpheme_t.c
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
	int real_length = 0;

	if (one.i >= two.i) {
		longest = one;
		shortest = two;
	} else {
		longest = two;
		shortest = one;
	}

	char *str = malloc(sizeof(char) * (shortest.i + 1));
	while (shortest.word[real_length] != '\0' && shortest.word[real_length] == longest.word[real_length]) {
		str[real_length] = shortest.word[real_length];
		++real_length;
	}

	str[real_length] = '\0';
	str = realloc(str, (sizeof(char) * (real_length+1)));
	if (str == NULL)
		exit(E_REALLOC);

	char *regex = malloc(sizeof(char) * (real_length + 2));
	regex[0] = '^';
	strcat(regex, str);
	struct morpheme_t longest_match;

	longest_match.regex = regex;
	longest_match.freq = 1;
	longest_match.morpheme = str;
	
	free(str);
	return longest_match;

}
