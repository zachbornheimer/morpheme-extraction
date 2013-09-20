typedef struct ngram
{
	int mySize;
	member myMembers[mySize];
} ngram;

typedef struct member
{
	member subMembers[1];
	n_gram* myNgram;
	char* word;
	int freq;
	void expand();
} member;

void expand()
{
	subMembers = malloc(sizeof(subMembers + 1));
};

