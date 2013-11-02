/*
 * structs.h
 * Unify Struct Definitions in one file
 */

#ifndef STRUCT_DEFS
#define STRUCT_DEFS

struct char_doubleton {
	char c;
	int freq;
};

struct word_t {
	unsigned int i;
	int freq;
	char *word;
};

#endif
