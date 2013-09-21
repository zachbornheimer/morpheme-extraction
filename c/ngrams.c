/*
File: ngrams.c
Author: Kyle Ashley
Purpose: Define Data Structures for n-grams
*/

#include <stdio.h>
#include <stdlib.h>
/*
	member = element x of the ngram,
	a member consists of an array of submembers
	each submember is defined as [ngram, word, freq]
	
	ngram = $size x {members}
	member = [
				[&ngram, word, freq]
				...
				[&ngram, word, freq]
			]
*/

//typedef struct ngram {} ngram;
//typedef struct member {} member;
//typedef struct submember {} submember;
struct member;

typedef struct ngram{
	int mySize;
	struct member* members;
} ngram;

typedef struct member {	
	struct submember* submembers;
} member;

typedef struct submember {
	ngram* myNgram;
	char* word;
	int freq;
} submember;

void expand(member* this)
{
	this->submembers = malloc((sizeof(this->submembers) + 1) * sizeof(this->submembers));
}

void word(char* string, submember* this)
{
		this->word = string;
		++this->freq;	
}


int main()
{
	int ngramSize = 9;
	ngram* newngram = malloc(sizeof(ngram));
	newngram->mySize = ngramSize;
	char var[] = "Hi";
	word(var, newngram->member->submembers[0]);
	printf("%s", newngram->members->submembers->word);
//	printf("%c", newngram->members->submembers->word);
//	printf(newngram->members->submembers->word);
	return 0;
}

