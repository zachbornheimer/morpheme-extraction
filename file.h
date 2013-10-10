/*
 * file.h
 * Purpose: Define everything for reading a file and processing it
 * Ex: 
 * main()
 * {
 *     file f = read_file(ask_for_file());
 *     printf("%s\n", f);
 * }
 */
 
#include <stdlib.h>
#include <stdio.h>

typedef char* file;

/* prototypes, for good measure */
void read_by_char(const int c, file *buffer, int *charCount, int *bytes);
file read_file(const char *filename);
file ask_for_file();
/* end prototypes */

/* adds a character to a file (char *array) type */
void read_by_char(const int c, file *buffer, int *charCount, int *bytes)
{
    if (*bytes <= *charCount) {
        *bytes = *charCount + 1; 
        *buffer = realloc(*buffer, sizeof(char) * *bytes);
        if(buffer == NULL) {
            --(*charCount);
            printf("Cannot reallocate memory for reading file!\n");
            exit(1);
        }
    }
    (*buffer)[(*charCount)++] = c;
    (*buffer)[(*charCount)] = '\0';
}

/* returns a file's contents in: file fileContents (char *fileContents) */
file read_file(const char *filename)
{
    int c, charCount = 0, bytes = 1;
    FILE *alpha;

    file f = malloc(sizeof(char) * bytes);

    if((alpha = fopen(filename, "r")) == NULL)
        printf("Couldn't read file: %s", filename);

    while((c = fgetc(alpha)) != EOF)
        read_by_char(c, &f, &charCount, &bytes);

    fclose(alpha);

    return f;
}

/* user prompt for filename */
file ask_for_file()
{
    int c, charCount = 0, bytes = 1;
    char *filename;

    filename  = malloc(sizeof(char) * bytes);

    while ((c = getchar()) != EOF && c != '\n')
        read_by_char(c, &filename, &charCount, &bytes);

    return filename;
}
