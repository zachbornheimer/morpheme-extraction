/*
 * ngram_t.h
 * Define the prototypes for ngram_t
 */

#ifndef NGRAM_T_PROTO
#define NGRAM_T_PROTO

/* for ngram_t */
//int add_ngram_element(struct ngram_t*, int);
int add_ngram_element(struct word_t_array**, int);
struct ngram_t new_ngram(void);
void free_ngram(struct ngram_t*);
void quicksort_ngram_array(struct ngram_t***, int, int);
void uniqify(struct ngram_half_array*);
int ngrams_similar(struct ngram_t, struct ngram_t);
void add_similar_ngram_ref(struct ngram_t**, struct ngram_t**);

#endif
