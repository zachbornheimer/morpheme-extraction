/*
 * file.h
 * Define the prototypes for file.c
 */

#ifndef FILE_PROTO
#define FILE_PROTO

/* for file reading and whatnot */
void read_by_char(const int, char**, int*, int*);
char* read_file(const char*);
char* ask_for_file(void);

#endif
