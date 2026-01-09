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
#include "alphabet.h"


/* retrieves word delimiter */
char* find_word_delimiter(char **file)
{
	printf("DEBUG: find_word_delimiter started, file=%p\n", (void*)file);
	fflush(stdout);
	char *ret, **arr;
	V_PRINT("Creating List of Uniq Chars.");
	printf("DEBUG: Calling uniq...\n");
	fflush(stdout);
	int j, i = uniq(file, &ret);
	printf("DEBUG: uniq returned, i=%d\n", i);
	fflush(stdout);
	if (i == -1) {
		errno = E_UNIQ;
		return "";
	}
	V_PRINT("Constructing Word Delimiter Structures.");
	printf("DEBUG: About to create char_doubleton array of size %d\n", i);
	fflush(stdout);
	struct char_doubleton wd[i];
	printf("DEBUG: char_doubleton array created\n");
	fflush(stdout);
	printf("DEBUG: Starting loop j=0 to %d (should be < not <=)\n", i);
	fflush(stdout);
	for (j = 0; j < i; ++j) {
		if (j < 3) {
			printf("DEBUG: Loop iteration j=%d\n", j);
			fflush(stdout);
		}
		if (verbose_mode == ON)
			printf("Working on character %d/%d...\r", j, i);
		const char c[2] = {ret[j], '\0'};
		if (j < 3) {
			printf("DEBUG: Calling explode_sansnull for char '%c'\n", ret[j]);
			fflush(stdout);
		}
		int freq = explode_sansnull(&arr, *file, c);
		if (j < 3) {
			printf("DEBUG: explode_sansnull returned freq=%d\n", freq);
			fflush(stdout);
		}
		wd[j].c = ret[j];
		wd[j].freq = freq;
		free(arr);
	}
	printf("DEBUG: Loop finished\n");
	fflush(stdout);
	printf("DEBUG: Freeing ret...\n");
	fflush(stdout);
	free(ret);
	printf("DEBUG: ret freed\n");
	fflush(stdout);
	V_PRINT("Initial Word Delimiter Structures Constructed.");

	printf("DEBUG: Allocating wd_real...\n");
	fflush(stdout);
	int max_size = 0, index = 0;
	char *wd_real = malloc(sizeof(char) * 2);
	printf("DEBUG: wd_real allocated\n");
	fflush(stdout);

	printf("DEBUG: Starting second loop j=0 to %d (should be < not <=)\n", i);
	fflush(stdout);
	for (j = 0; j < i; ++j) {
		if (j < 3 || j >= i-1) {
			printf("DEBUG: Second loop iteration j=%d, wd[j].freq=%d, max_size=%d\n", j, wd[j].freq, max_size);
			fflush(stdout);
		}
		if (wd[j].freq > max_size) {
			if (j < 3) {
				printf("DEBUG: Found new max at j=%d\n", j);
				fflush(stdout);
			}
			index = 0;
			wd_real[index] = wd[j].c;
			wd_real[++index] = '\0';
			max_size = wd[j].freq;
		} 
		else if (wd[j].freq == max_size && !in_array((wd[j].c), &wd_real, index)) {
			printf("DEBUG: Equal max at j=%d\n", j);
			fflush(stdout);
			wd_real[index] = wd[j].c;
			wd_real[++index] = '\0';
			wd_real = realloc(wd_real, sizeof(char) * index+2);
			REALLOC_CHECK(wd_real);
			char *uniqd;
			uniq(&wd_real, &uniqd);
			free(wd_real);
			wd_real = uniqd;
		}
	}
	printf("DEBUG: Second loop finished, max_size=%d\n", max_size);
	fflush(stdout);
	V_PRINT("Word Delimiter Candidates Identified.");

	printf("DEBUG: wd_real='%s', len=%zu\n", wd_real, strlen(wd_real));
	fflush(stdout);
	char *wd_final;
	char *wd_permuted = malloc(strlen(wd_real) * sizeof(char) + 1);
	strcpy(wd_permuted, wd_real);
	int wd_final_freq = 0, size;
	i = 0;
	printf("DEBUG: Checking if strlen(wd_real) > 1: %d\n", strlen(wd_real) > 1);
	fflush(stdout);
	if(strlen(wd_real) > 1) {
		printf("DEBUG: Entering permutation while loop\n");
		fflush(stdout);
		while(strcmp(wd_real, permute(&wd_permuted, &i)) != 0) {
			printf("DEBUG: Permutation iteration i=%d\n", i);
			fflush(stdout);
			const char *wd_testing = wd_permuted;
			if ((size = explode_sansnull_str(&arr, *file, &wd_testing)) > wd_final_freq) {
				size = wd_final_freq;
				wd_final = wd_permuted; 
			}  else if (size == wd_final_freq && size != 0) {
				printf("This implementation is not designed to handle multiple word delimiters.\nSkipping this file...\n");
				errno = E_OVERRULED;
			}
		}
		printf("DEBUG: Exited permutation loop\n");
		fflush(stdout);
	}

	free(wd_permuted);
	
	printf("DEBUG: About to return, wd_final_freq=%d\n", wd_final_freq);
	fflush(stdout);
	if (wd_final_freq == 0) {
		printf("DEBUG: Returning wd_real='%s'\n", wd_real);
		fflush(stdout);
		return wd_real;
	}
	else {
		printf("DEBUG: Returning wd_final='%s'\n", wd_final);
		fflush(stdout);
		return wd_final;
	}
}

#endif
