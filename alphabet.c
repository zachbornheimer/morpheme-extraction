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

#include "constants.h"
#include "prototypes.h"

char* find_word_delimiter(char**);

int main(void)
{
	char *f;
	int index = 0;
	char *wd;
	while((f = getfiles(&index))) {
		if (f != NULL && strlen(f) > 2)
			wd = find_word_delimiter(&f);
		printf("%s\n", wd);
		free(f);
	}

	return 0;
}

char* find_word_delimiter(char **f)
{
	char *ret, **arr;
	int j, i = uniq(f, &ret);
	struct char_doubleton wd[i];
	for (j = 0; j <= i; ++j) {
		int freq = explode_sansnull(&arr, *f, ret[j]);
		wd[j].c = ret[j];
		wd[j].freq = freq;
		free(arr);
	}
	free(ret);

	char *wd_real = malloc(sizeof(char) + 1);
	int max_size = 0, index = 0;

	for (j = 0; j <= i; ++j) {
		if (wd[j].freq > max_size) {
			free(wd_real);
			wd_real = malloc(sizeof(char) + 1);
			wd_real[0] = wd[j].c;
			wd_real[1] = '\0';
			max_size = wd[j].freq;
		} 
		else if (wd[j].freq == max_size) {
			wd_real[index++] = wd[j].c;
			wd_real[index] = '\0';
			wd_real = realloc(wd_real, sizeof(char) * index+1);
		}
	}

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
			}  else if (size == wd_final_freq && wd_final != wd_real) {
				printf("This implementation is not designed to handle multiple word delimiters.\nSkipping this file...\n");
				errno = EOVERRULED;
			}
		}

	free(wd_permuted);

	if (wd_final_freq == 0)
		return wd_real;
	else
		return wd_final;
}

#endif
