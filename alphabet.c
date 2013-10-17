/* alphabet.c
 * Implement the Alphabet Extraction Mechanism
 *   for a particular file
 */

#include "functions.h"
#include <strings.h>


/*
 *  The goal implementation is to be able to use the getfiles as a loop:
 *  while((char *f = getfiles()) {
 *  	    process(f);
 *  	    free(f);
 *  }
 */

int main(void)
{

/*
 * Currently, I need to implement a hash (doubleton)
 * which will store the character frequencies that way it
 * can check if there are multiple characters in the delimiter.
 *
 * Also, I need to implement the mechanism for handling permutations
 * for discerning multicharacter word delimiters.
 */

	char *f;
	int index = 0;
	char *wd;
	while((f = getfiles(&index))) {
		struct char_doubleton *word_delimiter;
		if (f != NULL && strlen(f) > 2)
			wd = find_word_delimiter(&f);
		printf("%s\n", wd);
		free(f);
	}
}


