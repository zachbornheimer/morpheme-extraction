// Package file provides file I/O operations from file.c
// Line-for-line translation of file.c lines 1-96
package file

import (
	"fmt"
	"io"
	"os"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/filesystem"
	"github.com/zachbornheimer/morpheme-extraction/internal/structs"
)

// DefaultFS is the default filesystem implementation
// Can be overridden for testing
var DefaultFS filesystem.FileSystem = filesystem.NewOsFileSystem()

// ReadByChar adds a character to buffer (lines 23-32)
func ReadByChar(c byte, buffer *[]byte, charCount *int, bytes *int) {
	if *bytes <= *charCount+1 {
		*bytes = *charCount + 2
		newBuffer := make([]byte, *bytes)
		copy(newBuffer, *buffer)
		*buffer = newBuffer
	}
	(*buffer)[*charCount] = c
	*charCount++
	(*buffer)[*charCount] = 0
}

// ReadFile returns a file's contents (lines 35-52)
func ReadFile(filename string) string {
	charCount := 0
	bytes := 1
	f := make([]byte, bytes)

	alpha, err := DefaultFS.Open(filename)
	if err != nil {
		fmt.Printf("Couldn't read file: %s\n", filename)
		return ""
	}
	defer alpha.Close()

	// Read file byte by byte
	buf := make([]byte, 1)
	for {
		n, err := alpha.Read(buf)
		if err == io.EOF {
			break
		}
		if n > 0 {
			ReadByChar(buf[0], &f, &charCount, &bytes)
		}
	}

	// Return as string, trimming null bytes
	return string(f[:charCount])
}

// AskForFile prompts user for filename (lines 55-66)
func AskForFile() string {
	charCount := 0
	bytes := 1
	filename := make([]byte, bytes)

	stdin := DefaultFS.Stdin()
	buf := make([]byte, 1)
	for {
		n, err := stdin.Read(buf)
		if err == io.EOF || (n > 0 && buf[0] == '\n') {
			break
		}
		if n > 0 {
			ReadByChar(buf[0], &filename, &charCount, &bytes)
		}
	}

	return string(filename[:charCount])
}

// WriteToFile writes lexical categories to output file (lines 68-93)
func WriteToFile(filename string, lex []structs.LexicalCategories, count int, header string) {
	fmt.Printf("Writing data to file: %s\n", filename)

	// Open file in append mode
	output, err := DefaultFS.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not open file, %s, instead writing to %s\n", filename, constants.ZedramOutput)
		output, err = DefaultFS.OpenFile(constants.ZedramOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Something catastrophic occurred. Run again -- maybe redirect the output?\n")
			return
		}
	}
	defer output.Close()

	// Write header
	fmt.Fprintf(output, "\n%s\n", header)
	for i := 0; i < len(header); i++ {
		fmt.Fprintf(output, "=")
	}
	fmt.Fprintf(output, "\n")

	// Write morphemes
	for i := 0; i < count; i++ {
		if lex[i].Morpheme != nil && lex[i].Morpheme.Type != constants.UNDEF {
			fmt.Fprintf(output, "%s: %s \n", constants.StrType(lex[i].Morpheme.Type), lex[i].Morpheme.Morpheme)
		}
	}
}

