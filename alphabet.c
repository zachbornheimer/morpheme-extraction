/* alphabet.c
 * Implement the Alphabet Extraction Mechanism
 *   for a particular file
 */

#include "file.h"
#include "functions.h"
#include <dirent.h>
#include <strings.h>

#define DEFAULT_PATH "."

/*
 
 The goal implementation is to be able to use the getfiles as a loop:
    while((file f = getfiles()) {
       process(f);
       free(f);
    }
 
 
For now, because of the limits of my knowledge, I am omitting recursion.
 Line 63 starts the recursion, but it fails.  I'm fairly certain that I'm polluting the namespace
    but it isn't a good use of my time currently to investigate that.
 
 */

file __gf(char dirpath[], int *index, int *count)
{
    DIR *currdir;
    char *path = malloc(strlen(dirpath) + 2);
    char *terminate = "/";
    struct dirent *d;
    strcpy(path, dirpath);
    int p = strlen(path);  /* does not include \0 */
    int bytes = p;
    if (!path)
        path = ".";

    if (path[p-1] != '/') {
        if (path == NULL)
            return NULL;
        strcat(path, terminate);
    }
    currdir = opendir(path);
    file f;
    if (currdir)
        while ((d = readdir(currdir)) != NULL)
            if (d->d_type == DT_REG && d->d_name[0] != '.') {
                char *curfile = append(path, d->d_name);
               // printf("%s %d %d\n", curfile, *index, *count);
                if ((*count)++ >= *index) {
                    ++(*index);
                    f = read_file(curfile);
                    return curfile;
                    free(curfile);
                    
                }
            } else {
                if (d->d_type == DT_DIR && d->d_name[0] != '.') {
                    char *cfile = append(path, d->d_name);
                    file f;
                    __gf(cfile, index, count);
                    free(cfile);
                }
            }
    else
        return NULL;

      return NULL; /* to stifle the compiler's error checking */

  

}

file getfiles(int *index)
{
    /* index is how many files it should parse to catch up to where it was in the previous run */
    /* count is the current index num */
    int count = 0;
    return __gf(".", index, &count);
}

int main(void) {
    file f;
    int index = 0;
    while((f = getfiles(&index))) {
       printf("%s\n", f);
    }
    free(f);
}
