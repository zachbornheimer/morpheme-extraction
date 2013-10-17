/*
 * file.h
 * Purpose: Define everything for reading a file and processing it
 * Ex: 
 * main()
 * {
 *     char *f = read_file(ask_for_file());
 *     printf("%s\n", f);
 * }
 */

#ifndef FILE_FUNCTIONS
#define FILE_FUNCTIONS

#include "prototypes.h"
#include <stdlib.h>
#include <stdio.h>

/* adds a character to a file (char *array) type */
void read_by_char(const int c, char **buffer, int *charCount, int *bytes)
{
	if (*bytes <= *charCount + 1) {
		*bytes = *charCount + 2; 
		*buffer = realloc(*buffer, sizeof(char) * *bytes);
		if(buffer == NULL) {
			--(*charCount);
			printf("Cannot realloc memory for reading file!\n");
			exit(1);
		}
	}
	(*buffer)[(*charCount)++] = c;
	(*buffer)[(*charCount)] = '\0';
}

/* returns a file's contents in: char *fileContents */
char* read_file(const char *filename)
{
	int c, charCount = 0, bytes = 1;
	FILE *alpha;
	char *f = malloc(sizeof(char) * bytes);

	if ((alpha = fopen(filename, "r")) == NULL) {
		printf("Couldn't read file: %s\n", filename);
		fclose(alpha);
		return NULL;
	} else {
		while((c = fgetc(alpha)) != EOF)
			read_by_char(c, &f, &charCount, &bytes);
	}

	fclose(alpha);
	return f;
}

/* user prompt for filename */
char *ask_for_file()
{
	int c, charCount = 0, bytes = 1;
	char *filename;

	filename = malloc(sizeof(char) * bytes);

	while ((c = getchar()) != EOF && c != '\n')
		read_by_char(c, &filename, &charCount, &bytes);

	return filename;
}

#endif
