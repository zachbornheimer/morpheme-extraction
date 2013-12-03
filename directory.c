/*
 * directory.c
 * Get files in a directory
 * by Z. Bornheimer
 */

#ifndef DIRECTORY
#define DIRECTORY

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <errno.h>

#include "constants.h"
#include "functions.h"
#include "file.h"
#include "directory.h"

char* getfiles(int *index)
{
	/* index is how many files it should parse to catch up to where it was in the previous run */
	/* count is the current index num */
	int count = 0;
	return __gf(DEFAULT_PATH, index, &count);
}

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
			if ((d->d_type == DT_REG || d->d_type == DT_UNKNOWN) && d->d_name[0] != '.') {
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

#endif
