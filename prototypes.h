#ifndef PROTOTYPES
#define PROTOTYPES

#include "structs.h"

void assemble_structures();
void alphabet_identification();
void morpheme_extraction();
void lexical_category_identification();
void syntax_model_creation();
void meaning_map_generation();

/* for file reading and whatnot */
void read_by_char(const int, char**, int*, int*);
char* read_file(const char*);
char* ask_for_file();

/* for alphabet extraction */
char* __gf(char[], int*, int*);
char* getfiles(int*);
char* find_word_delimiter(char**);

/* for words_t */
void setword(struct words_t*, char*);

/* general functions */
char* append(char*, char*);
void expand(char**);
int in_array(const int, char**);
int uniq(char**, char**);
int explode_sansnull(char***, char*, char);
int explode_sansnull_str(char***, char*, char**);

char* permute(char **string, int *i);
int move_char(int *index, char **in);

#endif
