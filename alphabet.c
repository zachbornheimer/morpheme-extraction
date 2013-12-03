#ifndef ALPHABET
#define ALPHABET

/* alphabet.c
 * Implement the Alphabet Extraction Mechanism
 *   for a particular file
 */


/*
 *  The goal implementation is to be able to use the getfiles as a loop:
 *  while((char *f = getfiles()) {
 *  	    process(f);
 *  	    free(f);
 *  }
 */

#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>

#include "structs.h"
#include "constants.h"
#include "functions.h"

char* find_word_delimiter(char**);

char* find_word_delimiter(char **f)
{
	char *ret, **arr;
	V_PRINT("Creating List of Uniq Chars.");
	int j, i = uniq(f, &ret);
	if (i == -1) {
		errno = E_UNIQ;
		return "";
	}
	V_PRINT("Constructing Word Delimiter Structures.");
	struct char_doubleton wd[i];
	for (j = 0; j <= i; ++j) {
		if (verbose_mode == ON)
			printf("Working on character %d/%d...\r", j, i);
		char c[2] = {ret[j], '\0'};
		int freq = explode_sansnull(&arr, *f, c);
		wd[j].c = ret[j];
		wd[j].freq = freq;
		//free(*arr);
		free(arr);
	}
	free(ret);
	V_PRINT("Initial Word Delimiter Structures Constructed.");

	int max_size = 0, index = 0;
	char *wd_real = malloc(sizeof(char) + index+1);

	for (j = 0; j <= i; ++j) {
		if (wd[j].freq > max_size) {
			//free(wd_real);
			index = 0;
			//wd_real = malloc(sizeof(char)+1);
			wd_real[index] = wd[j].c;
			wd_real[++index] = '\0';
			max_size = wd[j].freq;
		} 
		else if (wd[j].freq == max_size && !in_array((wd[j].c), &wd_real, index)) {
			wd_real[index] = wd[j].c;
			wd_real[++index] = '\0';
			wd_real = realloc(wd_real, sizeof(char) * index+2);
			if (wd_real == NULL) {
				printf("ERROR, DID NOT REALLOC\n");
				exit(E_REALLOC);
			}
			char *uniqd;
			uniq(&wd_real, &uniqd);
			free(wd_real);
			wd_real = uniqd;
		}
	}
	V_PRINT("Word Delimiter Candidates Identified.");

	char *wd_final;
	char *wd_permuted = malloc(strlen(wd_real) * sizeof(char) + 1);
	strcpy(wd_permuted, wd_real);
	int wd_final_freq = 0, size;
	i = 0;
	if(strlen(wd_real) > 1)
		while(strcmp(wd_real, permute(&wd_permuted, &i)) != 0) {
			if ((size = explode_sansnull_str(&arr, *f, &wd_permuted)) > wd_final_freq) {
				size = wd_final_freq;
				wd_final = wd_permuted; 
			}  else if (size == wd_final_freq) {
				printf("This implementation is not designed to handle multiple word delimiters.\nSkipping this file...\n");
				errno = E_OVERRULED;
			}
		}

	free(wd_permuted);

	if (wd_final_freq == 0)
		return wd_real;
	else
		return wd_final;
}

#endif
