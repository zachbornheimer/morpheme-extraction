/*
 * functions.h
 * Define the function prototypes
 */

#ifndef FUNCTION_PROTO
#define FUNCTION_PROTO

/* general functions */
char* append(char*, char*);
int in_array(const int, char**, const int);
int in_char_array(const char*, char*[], const int);
int uniq(char**, char**);
int uniq_words(char**, int);
int explode_sansnull_str(char***, const char*, const char**);
int explode_sansnull(char***, const char*, const char*);
int move_char(int*, char**);
char* permute(char**, int*);
char* reverse(char*);
void swap_mem(void*,void*);

#endif
