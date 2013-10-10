#include <stdio.h>
#include <string.h>

/* append two to the end of one */
char* append(char* one, char* two)
{
    char *temp = malloc(sizeof(char) * (strlen(one) + 2 + strlen(two)));
    strcpy(temp, one);
    strcat(temp, two);
    return temp;
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
