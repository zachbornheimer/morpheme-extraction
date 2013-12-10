/*
 * file.c
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

#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "file.h"
#include "morpheme_t.h"

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
char* ask_for_file(void)
{
	int c, charCount = 0, bytes = 1;
	char *filename;

	filename = malloc(sizeof(char) * bytes);

	while ((c = getchar()) != EOF && c != '\n')
		read_by_char(c, &filename, &charCount, &bytes);

	return filename;
}

void write_to_file(char *filename, struct lexical_categories_t *lex, int count, char* header)
{

	FILE *output;

	printf("Writing data to file: %s\n", filename);

	output = fopen(filename, "a");
	if (output == NULL) {
		printf("Could not open file, %s, instead writing to %s\n", filename, ZEDRAM_OUTPUT);
		output = fopen(ZEDRAM_OUTPUT, "a");
		if (output == NULL)
			printf("Something catastrophic occurred.  Run again -- maybe redirect the oputput?\n");
	}

	int i = 0;
	fprintf(output, "\n%s\n", header);
	for (i = 0; i < strlen(header); ++i)
		fprintf(output, "=");
	fprintf(output, "\n");
	for (i = 0; i < count; ++i)
		if (lex[i].morpheme.type != UNDEF)
			fprintf(output, "%s: %s\n", strtype(lex[i].morpheme.type), lex[i].morpheme.morpheme);

	fclose(output);
}

#endif
