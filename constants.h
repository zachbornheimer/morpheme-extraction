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

/* for switches */
#define ON  1
#define OFF 0

#define E_OVERRULED 8
#define E_REALLOC   9
#define E_MALLOC   10
#define E_UNIQ      6
#define E_ASSERT    7

/* thresholds - percents */
#define THRESHOLD_SIMILAR_NGRAMS 12.5
#define THRESHOLD_CIRCUMFIX       100
#define THRESHOLD_CONFIRMATION     10

/* Morpheme Types */
#define UNDEF     0
#define PREFIX    1
#define SUFFIX    2
#define CIRCUMFIX 3
#define INFIX     4
#define STEM      5


#include "externs.h"
#include "macros.h"

#endif
