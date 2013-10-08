/* alphabet.c
 * Implement the Alphabet Extraction Mechanism
 *   for a particular file
 */

#include "file.h"
#include "functions.h"
#include <dirent.h>
#include <strings.h>

getfiles(char* path)
{
    DIR *currdir;
    char* t = "/";
    struct dirent *d;
    char* p = path;
    if (!path)
        path = ".";
    printf("%i %s %c\n", strlen(p), p, p[(strlen(p))]);
    if (path[strlen(p)] != '/') {
        path = append(path, t);
       // path = realloc(path, sizeof(char) * (strlen(path)+2) );
   //     path[strlen(p)] = '/';
    }
    printf("%s\n\n", p);
    return 1;
    currdir = opendir(path);
    if (currdir) {
        while ((d = readdir(currdir)) != NULL) {
            if (d->d_type == DT_REG && d->d_name[0] != '.')
                printf("%s/%s\n", path, d->d_name);
            else
                if (d->d_type == DT_DIR && d->d_name[0] != '.') {
                    //printf("%s\n", d->d_name);
                    path = realloc(path, sizeof(char) * 5);
                    //path = append(path,p);
                   // printf("%s\n",p);
                    return 1;
                    getfiles(append(path, d->d_name));
                }
        }
    }
    /*file doc = read_file(ask_for_file());
    printf("%s", doc);
*/
}

main() { getfiles("test"); }