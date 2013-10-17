#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <errno.h>

#include "prototypes.h"
#include "file.h"
#include "data_types.h"
#include "constants.h"



char* __gf(char dirpath[], int *index, int *count)
{
	DIR *currdir;
	char *path;
	path = malloc(strlen(dirpath) + 2);
	char *terminate = "/";
	struct dirent *d;
	strcpy(path, dirpath);
	int p = strlen(path);  /* does not include \0 */
	if (!path)
		path = ".";
	if (path[p-1] != '/') {
		if (path == NULL)
			return NULL;
		strcat(path, terminate);
	}

	currdir = opendir(path);
	char *f;
	if (currdir) {
		/* 
		 * intentionally non-recursive
		 * as it would take too much time to implement properly
		 * as to not pollute the working *path
		 */
		while ((d = readdir(currdir)) != NULL)
			if (d->d_type == DT_REG && d->d_name[0] != '.') {
				char *curfile = append(path, d->d_name);
				if ((*count)++ >= *index) {
					++(*index);
					f = read_file(curfile);
					if (f != NULL) {
						printf("Reading: %s\n", curfile);
						if (curfile != NULL)
							free(curfile);
						
						closedir(currdir);
						if (path != NULL)
							free(path);
						return f;
					}
					if (curfile != NULL)
						free(curfile);
				}
			}
		closedir(currdir);
	} else {
		return NULL;
	}

	free(path);
	return NULL; /* to stifle the compiler's error checking */
}

char* getfiles(int *index)
{
	/* index is how many files it should parse to catch up to where it was in the previous run */
	/* count is the current index num */
	int count = 0;
	return __gf(DEFAULT_PATH, index, &count);
}

char* find_word_delimiter(char **f)
{
	char *ret, **arr;
	int j, i = uniq(f, &ret);
	struct char_doubleton wd[i];
	if (wd == NULL)
		return -1;
	for (j = 0; j <= i; ++j) {
		int freq = explode_sansnull(&arr, *f, ret[j]);
		wd[j].c = ret[j];
		wd[j].freq = freq;
		free(arr);
	}
	free(ret);

	char *wd_real = malloc(sizeof(char) + 1);
	int max_size = 0, index = 0;

	for (j = 0; j <= i; ++j) {
		if (wd[j].freq > max_size) {
			free(wd_real);
			wd_real = malloc(sizeof(char) + 1);
			wd_real[0] = wd[j].c;
			max_size = wd[j].freq;
		} 
		else if (wd[j].freq == max_size) {
			wd_real[index++] = wd[j].c;
			wd_real = realloc(wd_real, sizeof(char) * index+2);
		}
	}
	/*free(wd);*/

	char *wd_final;
	char *wd_permuted = malloc(strlen(wd_real) * sizeof(char) + 1);
	strcpy(wd_permuted, wd_real);
	int wd_final_freq = 0, size;
	i = 0;
	if(strlen(wd_real) > 1)
		while(!strcmp(wd_real, permute(&wd_permuted, &i)))
			if ((size = explode_sansnull(&arr, *f, wd_permuted)) > wd_final_freq) {
				size = wd_final_freq;
				wd_final = wd_permuted;
			} else if (size == wd_final_freq && wd_final != wd_real) {
				printf("This implementation is not designed to handle multiple word delimiters.\nSkipping this file...\n");
				errno = EOVERRULED;
			}

	/*free(wd_permuted); */

	if (wd_final_freq == 0)
		return wd_real;
	else
		return wd_final;
}

/* append two to the end of one */
char* append(char* one, char* two)
{
	char *temp = malloc(sizeof(char) * (strlen(one) + 2 + strlen(two)));
	if (temp == NULL)
		return "ERROR IN MALLOC";
	strcpy(temp, one);
	strcat(temp, two);
	return temp;
}

void expand(char **ptr)
{
	*ptr = realloc(*ptr, sizeof(char) * (strlen(*ptr) + 2));
}

int in_array(const int c, char **uniq)
{
	if (strlen(*uniq) == 0)
		return 0;
	for (int i = 0; i <= strlen(*uniq); ++i)
		if (c == (*uniq)[i])
			return 1;
	return 0;
}

int uniq(char **f, char **ret)
{
	if (f == NULL || *f == NULL)
		return -1;
	int len = strlen(*f) + 1;
	char *u = malloc(len * sizeof(char));
	int i = 0, j;
	for (j = 0; j < len; ++j)
		if ((*f) != NULL)
			if (j==0 || !in_array((*f)[j], &u))
				u[i++] = (*f)[j];

	*ret = u;
	return i-1;
}

/* function explode_sansnull taken from:
 * http://www.it.uu.se/katalog/larme597/explode
 *   used from public domain and modified by Z. Bornheimer
 */

int explode_sansnull(char ***arr_ptr, char *str, char delimiter)
{
	char *src = str, *end, *dst;
	char **arr;
	int size = 1, i, j = 0;

	while ((end = strchr(src, delimiter)) != NULL) {
		++size;
		src = end + 1;
	}

	arr = malloc(size * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

	src = str;
	dst = (char *) arr + size * sizeof(char *);
	for (i = 0; i < size; ++i) {
		if ((end = strchr(src, delimiter)) == NULL)
			end = src + strlen(src);
		if ((end-src) > 0)
			arr[j++] = dst;
		strncpy(dst, src, end - src);
		dst[end - src] = '\0';
		dst += end - src + 1;
		src = end + 1;
	}
	arr = realloc(arr, j * sizeof(char *) + (strlen(str) + 1) * sizeof(char));

	if (arr == NULL)
		exit(-1);
	*arr_ptr = arr;

	return j;
}

int move_char(int *index, char **in)
{
	if ((*in)[*index + 1] == '\0')
		*index = 0;
	int f = *index;
	char temp[strlen(*in)+1];
	int j;
	for (j = 0; j < f; ++j) {
		temp[j] = (*in)[j];
	}
	temp[j] = (*in)[j+1];
	temp[j+1] = (*in)[j];
	for (j = j+2; j <= strlen(*in); ++j)
		temp[j] = (*in)[j];
	temp[j] = '\0';
	strcpy(*in,temp);
	++(*index);
	return 1;
}

char* permute(char **string, int *i)
{
	if (*i >= strlen(*string))
		*i = 0;

	if (!move_char(i, string))
		printf("Could not move char.\n");

	return *string;
}

/*

#include "prototypes.h"
#include "morphemes.h"

void assemble_structures()
{
printf("%s", "Assemble Structures\n");
}

void alphabet_identification()
{
printf("%s", "Alphabet Identification\n");
}


void lexical_category_identification()
{
printf("%s", "Lexical Category Identification\n");
}

void syntax_model_creation()
{
printf("%s", "Syntax Model Creation\n");
}

void meaning_map_generation()
{
printf("%s", "Meaning Map Generation\n");
}

*/
