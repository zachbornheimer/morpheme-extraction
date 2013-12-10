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
#include "functions.h"

/* append two to the end of one */
char* append(char* one, char* two)
{
	char *temp = malloc(sizeof(char) * (strlen(one) + 1 + strlen(two)));
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

int in_array(const int c, char **uniq, const int size)
{
	int i;
	for (i = 0; i < size; ++i)
		if (c == (*uniq)[i])
			return i;
	return -1;
}

int in_char_array(const char *w, char *uniq[], const int size)
{
	int i;
	for (i = 0; i < size; ++i)
		if (strcmp(w, uniq[i]) == 0)
			return i;
	return -1;
}

/* returns base 0 */
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
			if (i == 0 || in_array((*f)[j], &u, i)==-1) {
				u[i++] = (*f)[j];
			}
		} else {
			break;
		}
	}
	*ret = u;
	return i-1;
}

/* returns base 0 */
int uniq_words(char **arr, int size)
{
	char **u = malloc(sizeof(char*) * size);
	int i = 0, j=0;
	for (j = 0; j < size; ++j) {
		/*if (verbose_mode == ON) {
			double p = ((double) j / (double) size) * (double) 100.00;
			printf("Creating Uniq Word Array %f%s Complete\r", p, "%");
		}*/
		if (i == 0 || in_char_array(arr[j], u, i)==-1) {
			u[i] = arr[j];
			++i;
		}
	}
	//free(u);
	return i-1;
}



int explode_sansnull_str(char ***arr_ptr, char *str, char **delimiter)
{
	int i, size = -1, len = strlen(*delimiter);
	for (i = 0; i < len; i++) {
		/*char *c = malloc(sizeof(char)*2);
		c[0] = delimiter[i];
		c[1] = '\0';*/
		size = explode_sansnull(arr_ptr, str, *delimiter);
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
	int wc, size = strlen(str)+1;
	int count = 0, char_count = 0, i = 0, j;
	char *token;

	char *s = strdup(str);
	char *t = strdup(delimiter);
	if (t == NULL)
		exit(E_REALLOC);

	for (j = 0; str[j] != '\0'; (str[j++] == *delimiter) ? ++(count) : 0);
	wc = count+1;
	count = 0;

	token = strtok(s, t);

	int malloc_size = (sizeof(char*) * (wc-1)) + size + wc;
	char **arr = malloc(malloc_size);
	
	while (token != NULL) {
		char_count += strlen(token)+1;
		//arr[i] = malloc(sizeof(char) * strlen(token)+1);
		if (token != NULL)
			arr[i++] = token;
		token = strtok(NULL, t);
		++count;
	}

	free(t);
	*a = arr;
	return i-1;
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

char* reverse(char* string)
{
	char temp;
	int i = 0, len = strlen(string)-1;

	if (len < 1)
		return string;

	char *str = string;

	while (i < (len-i)) {
		temp = str[i];
		str[i] = str[len-i];
		str[len-i] = temp;
		++i;
	}

	return str;
}

#endif
