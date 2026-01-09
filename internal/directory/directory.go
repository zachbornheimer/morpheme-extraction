// Package directory provides directory traversal from directory.c
// Line-for-line translation of directory.c lines 1-90
package directory

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/zachbornheimer/morpheme-extraction/internal/file"
	"github.com/zachbornheimer/morpheme-extraction/internal/functions"
)

// Global variable matching C (line 22)
var CorpusDir string

// GetFiles returns next file from corpus directory (lines 24-30)
func GetFiles(index *int, header *string) string {
	count := 0
	return getFilesInternal(CorpusDir, index, &count, header)
}

// getFilesInternal is the internal implementation (__gf in C, lines 32-88)
func getFilesInternal(dirpath string, index *int, count *int, header *string) string {
	path := dirpath
	
	// Ensure path ends with /
	if path == "" {
		path = "."
	}
	if len(path) > 0 && path[len(path)-1] != '/' {
		path = functions.Append(path, "/")
	}

	// Open directory
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Could not open dir: %s\n", path)
		return ""
	}

	// Iterate through directory entries
	for _, d := range entries {
		// Skip hidden files and directories
		if d.Name()[0] == '.' {
			continue
		}

		// Only process regular files
		info, err := d.Info()
		if err != nil {
			continue
		}

		if info.Mode().IsRegular() || info.Mode()&fs.ModeType == 0 {
			curfile := functions.Append(path, d.Name())
			
			if *count >= *index {
				(*index)++
				f := file.ReadFile(curfile)
				
				if f != "" {
					fmt.Printf("Reading: %s\n", curfile)
					*header = curfile
					return f
				}
			} else {
				*count++
			}
		}
	}

	return ""
}

