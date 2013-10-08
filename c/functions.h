#include <stdio.h>
#include <strings.h>

/* append two to the end of one */
char* append(char* one, char* two)
{
    int onesize, twosize, i;
    onesize = strlen(one);
    twosize = strlen(two);

    one = realloc(one, sizeof(char) * (onesize + twosize + 1));
    if (one == NULL)
        return -1;
    for (i = onesize + 1; (i-onesize-1) <= twosize; i++)
       one[i] = two[(i-onesize-1)]; 
    one[i] = '\0';
    return one;
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
