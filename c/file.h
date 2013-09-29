/*
 * file.h
 * by Z. Bornheimer
 * Purpose: Define everything for reading a file and processing it
 * Ex: file f = read_file(ask_for_file());
 */
 
#include <stdlib.h>
#include <stdio.h>

typedef char* file;

/* prototypes, for good measure */
void read_by_char(const int c, file *buffer, int *charCount);
file read_file(const char *filename);
file ask_for_file();
/* end prototypes */

/* adds a character to a file (char *array) type */
void read_by_char(const int c, file *buffer, int *charCount)
{
    (*buffer)[++(*charCount)] = '\0';
    *buffer = realloc(*buffer, sizeof(char) * *charCount);
    if(buffer == NULL) {
        printf("Cannot allocate memory for reading file!\n");
        exit(1);
    }
    (*buffer)[*charCount] = c;
}

/* returns a file's contents in: file fileContents (char *fileContents) */
file read_file(const char *filename)
{
    int c, charCount = -1;
    FILE *alpha;

    file f = malloc(sizeof(char));

    if((alpha = fopen(f, "r")) == NULL)
        printf("Couldn't read file: %s\n", f);

    while((c = fgetc(alpha)) != EOF)
        read_by_char(c, &f, &charCount);

    fclose(alpha);
    return f;
}

/* user prompt for filename */
file ask_for_file()
{
    int c, charCount = -1;
    char *filename;

    filename  = malloc(sizeof(char));

    while ((c = getchar()) != EOF && c != '\n')
        read_by_char(c, &filename, &charCount);

    return filename;
}
