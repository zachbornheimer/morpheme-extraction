/*
 * delimiter.c
 * Defines the word delimiter
 * by Z. Bornheimer
 */

#include "file.h"

int readfile(file *doc, int character);
void find_delimiter(file *doc, char *delimiter);

int main()
{
    int c;
    file doc;
    extern char *delimiter;

    find_delimiter(doc, delimiter);

    return 0;
}

int readfile(file doc, int character)
{
    doc->file[++doc->i] = character;
    return character;
}


