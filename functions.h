/*
 * functions.h
 * Define the function prototypes
 */

#ifndef FUNCTION_PROTO
#define FUNCTION_PROTO

/* general functions */
char* append(char*, char*);
void expand(char**);
int in_array(const int, char**, int);
int in_char_array(const char*, char*[], const int);
int uniq(char**, char**);
int uniq_words(char**, int);
int explode_sansnull_str(char***, char*, char**);
int explode_sansnull(char***, char*, char*);
int move_char(int *index, char **in);
char* permute(char **string, int *i);

#endif
