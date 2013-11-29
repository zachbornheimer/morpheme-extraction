/*
 * constants.h
 * Written by Z. Bornheimer (Zysys)
 *
 * Part of the Computerized Human Language Acquisition System
 */

#ifndef CONSTANTS
#define CONSTANTS

#define CORPUS_DIR "corpus/corpus"
#define ZEDRAM_OUTPUT "languageModel.zdrm"
//#define DEFAULT_PATH "./corpus/nlp-corpus/dev/normalizers/normalized/"
#define DEFAULT_PATH "./test-corpus/"
#define NGRAM_SIZE 9

#define BEFORE 0
#define AFTER 0

#define ON 1
#define OFF 0

#define E_OVERRULED 8
#define E_REALLOC 9
#define E_UNIQ 6

/* quicksort */
#define DESCENDING 0
#define ASCENDING 1

/* thresholds */
#define THRESHOLD_SIMILAR_NGRAMS 50


#include "externs.h"
#include "macros.h"

#endif
