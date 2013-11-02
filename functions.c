/*
 * functions.h
 * Contains all global-like functions for the
 *   ai system - this allows for non-specific things
 *   to be included without problems.
 *
 * Written by Z. Bornheimer
 */

#ifndef FUNCTIONS
#define FUNCTIONS

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "constants.h"

/* general functions */
char* append(char*, char*);
void expand(char**);
int in_array(const int, char**);
int uniq(char**, char**);
int explode_sansnull(char***, char*, char);
int explode_sansnull_str(char***, char*, char**);

char* permute(char **string, int *i);
int move_char(int *index, char **in);

/* append two to the end of one */
char* append(char* one, char* two)
{
	char *temp = malloc(sizeof(char) * (strlen(one) + 2 + strlen(two)));
	if (temp == NULL)
		return "ERROR IN MALLOC";
	strcpy(temp, one);
	strcat(temp, two);
	return temp;
}

void expand(char **ptr)
{
	*ptr = realloc(*ptr, sizeof(char) * (strlen(*ptr) + 2));
}

int in_array(const int c, char **uniq)
{
	int i;
	if (strlen(*uniq) == 0)
		return 0;
	for (i = 0; i <= (int) strlen(*uniq); ++i)
		if (c == (*uniq)[i])
			return 1;
	return 0;
}

int uniq(char **f, char **ret)
{
	if (f == NULL || *f == NULL)
		return -1;
	int len = strlen(*f);
	char *u = malloc(len * sizeof(char));
	int i = 0, j=0;
	for (j = 0; j <= len; ++j)
		if ((*f) != NULL)
			if (j==0 || !in_array((*f)[j], &u))
				u[i++] = (*f)[j];
	*ret = u;
	return i-1;
}

int explode_sansnull_str(char ***arr_ptr, char *str, char **delimiter)
{
	int i, size = -1, len = strlen(*delimiter);
	for (i = 0; i < len; i++)
		size = explode_sansnull(arr_ptr, str, (*delimiter)[i]);
	return size;
}

/* 
 * function explode_sansnull taken from:
 * http://www.it.uu.se/katalog/larme597/explode
 *   used from public domain and modified by Z. Bornheimer
 */

int explode_sansnull(char ***arr_ptr, char *str, char delimiter)
{
	char *src = str, *end, *dst;
	char **arr;
	int size = 1, i, j = 0;

	while ((end = strchr(src, delimiter)) != NULL) {
		++size;
		src = end + 1;
	}

	arr = malloc(size * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

	src = str;
	dst = (char *) arr + size * sizeof(char *);
	for (i = 0; i < size; ++i) {
		if ((end = strchr(src, delimiter)) == NULL)
			end = src + strlen(src);
		if ((end-src) > 0)
			arr[j++] = dst;
		strncpy(dst, src, end - src);
		dst[end - src] = '\0';
		dst += end - src + 1;
		src = end + 1;
	}
	arr = realloc(arr, j * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

	if (arr == NULL)
		exit(-1);
	*arr_ptr = arr;

	return j;
}

int move_char(int *index, char **in)
{
	if ((*in)[*index + 1] == '\0')
		*index = 0;
	int j, f = *index;
	char temp[strlen(*in)+1];
	for (j = 0; j < f; ++j)
		temp[j] = (*in)[j];
	temp[j] = (*in)[j+1];
	temp[j+1] = (*in)[j];
	for (j = j+2; j <= (int) strlen(*in); ++j)
		temp[j] = (*in)[j];
	//temp[j] = '\0';
	strcpy(*in,temp);
	++(*index);
	return 1;
}

char* permute(char **string, int *i)
{
	if (*i >= (int) strlen(*string))
		*i = 0;

	if (!move_char(i, string))
		printf("Could not move char.\n");

	return *string;
}

/*

#include "prototypes.h"
#include "morphemes.h"

void assemble_structures()
{
printf("%s", "Assemble Structures\n");
}

void alphabet_identification()
{
printf("%s", "Alphabet Identification\n");
}


void lexical_category_identification()
{
printf("%s", "Lexical Category Identification\n");
}

void syntax_model_creation()
{
printf("%s", "Syntax Model Creation\n");
}

void meaning_map_generation()
{
printf("%s", "Meaning Map Generation\n");
}

*/

#endif
