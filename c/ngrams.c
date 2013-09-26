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

void expandSubmember(member* pthis)
{
    submember* temp;
    int size = sizeof(pthis->submembers) / sizeof(submember*);  // determines size of submember array
	temp = malloc(size + sizeof(submember*));                   //allocates 1 + sizeof submember array
	int i;
	for(i = 0; i < sizeof(pthis->submembers)/sizeof(submember*); i++)  // loops through index's of submember array
	{
	    temp[i] = pthis->submembers[i];     // copies data into temp
	}
    free(pthis->submembers);          // free old array;
    pthis->submembers = temp;
}

//***********************************************************************************
// Not sure if this is neccesary to the DS.
//
// Similar Functionality to 'expandSubmember' for the array of members of the n-gram
// **********************************************************************************
void expandMember(ngram* pthis)
{
    member* temp;
    int size = sizeof(pthis->members) / sizeof(member*);    // determines size of submember array
	temp = malloc(size + sizeof(member*));                  //allocates 1 + sizeof member array
	int i;
	for(i = 0; i < sizeof(pthis->members)/sizeof(member*); i++)  // loops through index's of member array
	{
	    temp[i] = pthis->members[i];     // copies data into temp
	}
    free(pthis->members);          // free old array;
    pthis->members = temp;
}

void word(char* string, submember* pthis)
{
		pthis->word = string;
}

int main()
{
//	int ngramSize = 9;
//	ngram* newngram = malloc(sizeof(ngram));
//	newngram->mySize = ngramSize;
//	char var[] = "Hi";
//	word(var, newngram->member->submembers[0]);
//	printf("%s", newngram->members->submembers->word);
//	printf("%c", newngram->members->submembers->word);
//	printf(newngram->members->submembers->word);
	return 0;
}

