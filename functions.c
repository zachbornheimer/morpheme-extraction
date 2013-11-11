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
int explode_sansnull(char***, char*, char*);
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
//	if (strcmp(*uniq, "") || (*uniq)[0] == 0 || (*uniq)[0] == 0)
//		return 0;
	for (i = 0; i < strlen(*uniq); ++i)
		if (c == (*uniq)[i])
			return 1;
	return 0;
}

int uniq(char **f, char **ret)
{
	if (f == NULL || *f == NULL)
		return -1;
	V_PRINT("Calculating strlen of file");
	int len = strlen(*f);
	char *u = malloc((len+2) * sizeof(char));
	int i = 0, j=0;
	for (j = 0; j <= len; ++j) {
		if (verbose_mode == ON) {
			double p = ((double) j / (double) len) * (double) 100.00;
			printf("Creating Uniq Array %f%s Complete\r", p, "%");
		}
		if ((*f)[j] != 0) {
			if (i == 0 || !in_array((*f)[j], &u)) {

				u[i++] = (*f)[j];
			}
		}
	}
	strcat(u, "\0");
	*ret = u;
	return i-1;
}

int explode_sansnull_str(char ***arr_ptr, char *str, char **delimiter)
{
	int i, size = -1, len = strlen(*delimiter);
	for (i = 0; i < len; i++) {
		char *c = malloc(sizeof(char) * 2);
		strcat(c, delimiter[i]);
		strcat(c, "\0");
		size = explode_sansnull(arr_ptr, str, c);
	}
	return size;
}

/* 
 * function explode_sansnull taken from:
 * http://www.it.uu.se/katalog/larme597/explode
 *   used from public domain and modified by Z. Bornheimer
 */

int explode_sansnull(char ***a, char *str, char *delimiter)
{
	int wc, size = strlen(str) + 1;
	int count = 0, char_count = 0, i = 0, j;
	char *token;

	char *s = strdup(str);
	char *t = strdup(delimiter);

	for (j = 0; str[j] != '\0'; (str[j++] == *delimiter) ? ++(count) : 0);
	wc = count+1;
	count = 0;

	token = strtok(s, t);

	int malloc_size = (sizeof(char*) * wc) + size + wc;
	//printf("%d\n", malloc_size);
	char **arr = malloc(malloc_size);
	
	while (token != NULL) {
		char_count += strlen(token)+1;
		//arr[i] = malloc(sizeof(char) * strlen(token)+1);
		arr[i++] = token;
		token = strtok(NULL, t);
		++count;
	}

	*a = arr;
	return i-1;
}
/*
int main()
{
	int count;
       	char *str, del, **arr;
//	del = malloc(sizeof(char) * (strlen("test")+1));
	del = '-';
	str = "This-is-an-interesting-example-of-strtok.";
	explode(&count, &arr, str, &del);
	printf("'%s'\n", *arr);
	printf("Words: %d, Sentence: %s\n", count, arr[0]);
	free(*arr);
	free(arr);
	return 0;
}*/

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
