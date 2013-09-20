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
	char word[1];
	int freq;
} submember;

void expand(member* pthis)
{
	pthis->submembers = malloc((sizeof(pthis->submembers) + 1) * sizeof(pthis->submembers));
};

int main()
{
	int ngramSize = 9;
	ngram newngram = {ngramSize};
	printf("%d", sizeof(newngram));
	printf("%d", sizeof(newngram.members->submembers->word));
	newngram.members->submembers->word[0] = 'c';
//	printf("%c", newngram.members->submembers->word);
//	printf(newngram.members->submembers->word);
	return 0;
}

