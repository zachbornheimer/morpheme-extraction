#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "file.h"

/* append two to the end of one */
char* append(char* one, char* two)
{
    char *temp = malloc(sizeof(char) * (strlen(one) + 2 + strlen(two)));
    strcpy(temp, one);
    strcat(temp, two);
    return temp;
}

void expand(char **ptr)
{
    *ptr = realloc(*ptr, sizeof(char) * (strlen(*ptr) + 2));
}

int in_array(const int c, char **uniq)
{
    for (int i = 0; i <= strlen(*uniq); ++i) {
        if (c == (*uniq)[i])
            return 1;
    }
    return 0;
}

int uniq(file *f, char **ret)
{
    int len = strlen(*f) + 1;
    char *uniq = malloc(len * sizeof(char));
    int i = 0, j;
    for (j = 0; j < len; ++j)
        if (!in_array((*f)[j], &uniq))
            uniq[i++] = (*f)[j];

    *ret = uniq;
    return i-1;
}

/* function explode_sansnull taken from:
 * http://www.it.uu.se/katalog/larme597/explode
 *   used from public domain and modified by Z. Bornheimer
 */

int explode_sansnull(char ***arr_ptr, char *str, char delimiter)
{
    char *src = str, *end, *dst;
    char **arr;
    int size = 1, i, j = 0;

    while ((end = strchr(src, delimiter)) != NULL) {
        ++size;
        src = end + 1;
    }

    arr = malloc(size * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

    src = str;
    dst = (char *) arr + size * sizeof(char *);
    for (i = 0; i < size; ++i) {
        if ((end = strchr(src, delimiter)) == NULL)
            end = src + strlen(src);
        if ((end-src) > 0)
            arr[j++] = dst;
        strncpy(dst, src, end - src);
        dst[end - src] = '\0';
        dst += end - src + 1;
        src = end + 1;
    }
    arr = realloc(arr, j * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

    if (arr == NULL)
        exit(-1);

    *arr_ptr = arr;

    return j;
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
