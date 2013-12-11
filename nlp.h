/*
 * nlp.h
 * Defines the prototypes for nlp.c
 */

int find_morphemes(struct ngram_t**, int, char*, struct lexical_categories_t**);
void build_ngram_relationships(char*, char*, int*, struct ngram_t***);
int main(const int, char*[]);
int nlp(void);
