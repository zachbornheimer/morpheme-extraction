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

/* for words_t */
void setword(struct words_t*, char*);

#endif
