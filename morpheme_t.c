/* * morpheme_t.c
 * The functions related to extracting morphemes
 */

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "structs.h"
#include "morpheme_t.h"
#include "morpheme_list_t.h"
#include "functions.h"

void add_word(struct morpheme_t *morpheme, struct word_t w)
{
	int size = sizeof(struct word_t) * ++(morpheme->words_count);
	morpheme->words = realloc(morpheme->words, size);
	morpheme->words[morpheme->words_count-1] = w;
}

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


void gen_regex(char *one, char *two, int place_one, int place_two, struct morpheme_t *morpheme)
{
	/* Step 1, Confirm Morpheme Placement */
	if (!(strlen(morpheme->morpheme) <= strlen(one) && strlen(morpheme->morpheme) <= strlen(two)))
		return;

	int i = 0;
	int size = sizeof(char) * (strlen(morpheme->morpheme) + 1);
	char *word1 = malloc(size), *word2 = malloc(size);
	for (i = place_one; i < (place_one + strlen(morpheme->morpheme)); ++i)
		word1[i - place_one] = one[i];
	word1[i - place_one] = '\0';
	for (i = place_two; i < (place_two + strlen(morpheme->morpheme)); ++i)
		word2[i - place_two] = two[i];
	word2[i - place_two] = '\0';
	if (!(strcmp(word1, word2) == 0 && strcmp(word1, morpheme->morpheme) == 0))
		return;

	/* Step 1.5, If Either Word = Morpheme, the morpheme->morpheme is a stem */

	morpheme->type = UNDEF;

	if (strcmp(one, morpheme->morpheme) == 0 || strcmp(two, morpheme->morpheme) == 0) {
		morpheme->regex = malloc(sizeof(morpheme->morpheme) + 3);
		morpheme->regex[0] = '^';
		morpheme->regex[1] = '\0';
		strcat(morpheme->regex, morpheme->morpheme);
		strcat(morpheme->regex, "$");
		morpheme->type = STEM;
		return;
	}

	/* Step 2, Extract Chars on Each Side */
	int size1 = sizeof(char) * place_one + 1;
	int size2 = sizeof(char) * (2+strlen(one) - ((strlen(morpheme->morpheme) + place_one)));
	int size3 = sizeof(char) * place_two + 1;
	int size4 = sizeof(char) * (2+strlen(two) - ((strlen(morpheme->morpheme) + place_two)));
	char *first_half1 = malloc(size1);
	char *second_half1 = malloc(size2);
	char *first_half2 = malloc(size3);
	char *second_half2 = malloc(size4);

	for (i = 0; i < place_one; ++i)
		first_half1[i] = one[i];
	first_half1[i] = '\0';

	for (i = (place_one + strlen(morpheme->morpheme)); i <= strlen(one); ++i)
		second_half1[i - (place_one + strlen(morpheme->morpheme))] = one[i];
	second_half1[i - (place_one + strlen(morpheme->morpheme))] = '\0';

	for (i = 0; i < place_two; ++i)
		first_half2[i] = two[i];
	first_half2[i] = '\0';

	for (i = (place_two + strlen(morpheme->morpheme)); i <= strlen(two); ++i)
		second_half2[i - (place_two + strlen(morpheme->morpheme))] = two[i];
	second_half2[i - (place_two + strlen(morpheme->morpheme))] = '\0';

	/* Step 3, Generate Regex for Each Side */
	int front_size = sizeof(char) * (strlen(first_half1) + strlen(first_half2));
	if (strlen(first_half1) > strlen(first_half2))
		front_size += sizeof(char) * ((strlen(first_half1) * 3) + 3);
	else
		front_size += sizeof(char) * ((strlen(first_half2) * 3) + 3);
	
	int back_size = sizeof(char) * (strlen(second_half1) + strlen(second_half2));
	if (strlen(second_half1) > strlen(second_half2))
		back_size += sizeof(char) * ((strlen(second_half1) * 3) + 3);
	else
		back_size += sizeof(char) * ((strlen(second_half2) * 3) + 3);

	front_size += 1+2; /* \0 + () */
	back_size += 1+2;
	
	char *front = malloc(front_size*2);
	char *back = malloc(back_size*2);


	front[0] = ')';
	front[1] = '\0';
	back[0] = ')';
	back[1] = '\0';
	i = 0;

	char *work = malloc(sizeof(char) * 3);
	work[0] = '\0';
	work[1] = '\0';
	first_half1 = reverse(first_half1);
	first_half2 = reverse(first_half2);
	morpheme->front_regex_arr = malloc(front_size*2);
	morpheme->front_regex_arr_index = -1;
	while (1) {
		if (i >= strlen(first_half1) && i >= (strlen(first_half2)))
			break;
		morpheme->front_regex_arr[i] = malloc(sizeof(char) * 3);
		morpheme->front_regex_arr[i][0] = '\0';
		strcat(front, "]\0");

		if (i < strlen(first_half2)) {
			work[0] = first_half2[i];
			strcat(front, work);
			strcat(morpheme->front_regex_arr[i], work);
		}

		if (i < strlen(first_half1)) {
			work[0] = first_half1[i];
			strcat(front, work);
			strcat(morpheme->front_regex_arr[i], work);
		}

		morpheme->front_regex_arr_index = i;

		strcat(front, "[\0");
		++i;
	}
	first_half1 = reverse(first_half1);
	first_half2 = reverse(first_half2);
	strcat(front, "(\0");
	front = reverse(front);

	work[0] = '\0';
	work[1] = '\0';
	second_half1 = reverse(second_half1);
	second_half2 = reverse(second_half2);
	i = 0;
	morpheme->back_regex_arr = malloc(back_size*4);
	
	morpheme->back_regex_arr_index = -1;
	while (1) {
		if (i >= strlen(second_half1) && i >= strlen(second_half2))
			break;
		strcat(back, "]\0");
		morpheme->back_regex_arr[i] = malloc(sizeof(char) * 3);
		morpheme->back_regex_arr[i][0] = '\0';
		if (i < strlen(second_half1)) {
			work[0] = second_half1[i];
			strcat(back, work);
			strcat(morpheme->back_regex_arr[i], work);
		}
		if (i < strlen(second_half2)) {
			work[0] = second_half2[i];
			strcat(back, work);
			strcat(morpheme->back_regex_arr[i], work);
		}
		morpheme->back_regex_arr_index = i;
		strcat(back, "[\0");
		++i;
	}
	second_half1 = reverse(second_half1);
	second_half2 = reverse(second_half2);
	strcat(back, "(\0");
	back = reverse(back);

	/* Step 4, Store Regex */

	morpheme->regex = malloc(sizeof(char) * (strlen(front) + strlen(back) + strlen(morpheme->morpheme) + 8));
	morpheme->regex[0] = '\0';
	morpheme->front_regex = malloc(sizeof(char) * (strlen(front) + 2));
	morpheme->front_regex[0] = '\0';
	morpheme->back_regex = malloc(sizeof(char) * (strlen(back) + 2));
	morpheme->back_regex[0] = '\0';
	front = realloc(front, sizeof(char) * (strlen(front) + 1));
	if (front == NULL)
		exit(E_REALLOC);
	back = realloc(back, sizeof(char) * (strlen(back) + 1));
	if (back == NULL)
		exit(E_REALLOC);
	strcat(morpheme->regex, front);
	strcat(morpheme->regex, morpheme->morpheme);
	strcat(morpheme->regex, back);
	strcat(morpheme->front_regex, front);
	strcat(morpheme->back_regex, back);
}

void merge_rules(struct morpheme_t *morpheme)
{
	int i = 0, front_size = 0, back_size = 0;
	for (i = 0; i <= morpheme->front_regex_arr_index; ++i)
		if (strlen(morpheme->front_regex_arr[i]) != NULL)
			front_size += strlen(morpheme->front_regex_arr[i]);
	++front_size;
	for (i = 0; i <= morpheme->back_regex_arr_index; ++i)
		if (strlen(morpheme->back_regex_arr[i]) != NULL)
			back_size += strlen(morpheme->back_regex_arr[i]);
	++back_size;
	char *front = malloc(front_size * 3);
	char *back = malloc(back_size * 3);
	front[0] = ')';
	front[1] = '\0';
	back[0] = ')';
	back[1] = '\0';
	i = 0;
	while (1) {
		if (i > morpheme->front_regex_arr_index)
			break;
		strcat(front, "]");
		char *arr;
		uniq(&(morpheme->front_regex_arr[i]), &arr);
		morpheme->front_regex_arr[i] = arr;
		strcat(front, morpheme->front_regex_arr[i]);
		strcat(front, "[");
		++i;
	}
	strcat(front, "(");
	front = reverse(front);

	i = 0;
	int back_index = 0;

	i = 0;
	while (1) {
		if (i > morpheme->back_regex_arr_index)
			break;
		strcat(back, "]");
		char *arr;
		uniq(&(morpheme->back_regex_arr[i]), &arr);
		morpheme->back_regex_arr[i] = arr;
		strcat(back, morpheme->back_regex_arr[i]);
		strcat(back, "[");
		++i;
	}
	strcat(back, "(");
	back = reverse(back);


	int size = 3;
	size += (int) strlen(front) + (int) strlen(back) + (int) strlen(morpheme->morpheme);
	morpheme->regex = malloc(sizeof(char) * size);
	morpheme->regex[0] = '\0';
	if (strcmp(morpheme->front_regex, "^") != 0) {
		morpheme->front_regex = malloc(sizeof(char) * (strlen(front) + 2));
		morpheme->front_regex[0] = '\0';
		strcat(morpheme->regex, front);
		strcat(morpheme->front_regex, front);
	} else {
		strcat(morpheme->regex, morpheme->front_regex);
	}
	strcat(morpheme->regex, morpheme->morpheme);
	if (strcmp(morpheme->back_regex, "$") != 0) {
		morpheme->back_regex = malloc(sizeof(char) * (strlen(back) + 2));
		morpheme->back_regex[0] = '\0';
		strcat(morpheme->regex, back);
		strcat(morpheme->back_regex, back);
	} else {
		strcat(morpheme->regex, morpheme->back_regex);
	}

}
