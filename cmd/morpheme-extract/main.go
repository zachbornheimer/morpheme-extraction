// main.go - Entry point for morpheme extraction system
// Line-for-line translation of nlp.c:main() lines 32-61
package main

import (
	"os"

	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/nlp"
)

// Global variables matching nlp.c lines 25-29
var (
	verboseMode         int
	processSequentially int
	processFull         int
	outputFilename      string
	corpusDir           string
)

func main() {
	argc := len(os.Args)
	argv := os.Args

	// Initialize globals (lines 35-39)
	verboseMode = constants.OFF
	processSequentially = constants.OFF
	processFull = constants.OFF
	outputFilename = constants.ZedramOutput
	corpusDir = constants.DefaultPath

	// Parse command line arguments (lines 40-58)
	// Note: C used OpenMP parallel for which caused race conditions, we use serial
	for i := 0; i < argc; i++ {
		if verboseMode == constants.OFF && argv[i] == "--verbose" {
			verboseMode = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--process-sequentially" {
			processSequentially = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--serial" {
			processSequentially = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--sequential" {
			processSequentially = constants.ON
		} else if processFull == constants.OFF && argv[i] == "--process" {
			processFull = constants.ON
		} else if argv[i] == "--output-file" {
			if i+1 < argc {
				outputFilename = argv[i+1]
			}
		} else if argv[i] == "--corpus-dir" {
			if i+1 < argc {
				corpusDir = argv[i+1]
			}
		}
	}

	// Call nlp() and return its exit code (line 60)
	exitCode := nlp.Nlp(verboseMode, processSequentially, processFull, outputFilename, corpusDir)
	os.Exit(exitCode)
}
