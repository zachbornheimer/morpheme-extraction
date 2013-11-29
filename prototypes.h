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

/* for word_t */
void setword(struct word_t*, char*);

/* for ngram_t */
//int add_ngram_element(struct ngram_t*, int);
int add_ngram_element(struct word_t_array**, int);
struct ngram_t new_ngram(void);
void free_ngram(struct ngram_t*);
void quicksort_ngram_array(struct ngram_t***, int, int);
void uniqify(struct ngram_half_array*);
int ngrams_similar(struct ngram_t, struct ngram_t);
void add_similar_ngram_ref(struct ngram_t**, struct ngram_t**);

/* general functions */
char* append(char*, char*);
void expand(char**);
int in_array(const int, char**, int);
int in_char_array(const char*, char*[], const int);
int uniq(char**, char**);
int uniq_words(char*[], int);
int explode_sansnull(char***, char*, char*);
int explode_sansnull_str(char***, char*, char**);

char* permute(char **string, int *i);
int move_char(int *index, char **in);

#endif
