// main.go - Entry point for morpheme extraction system
// Line-for-line translation of nlp.c:main() lines 32-61
package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

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
	cpuprofile          string
	memprofile          string
)

func main() {
	os.Exit(run())
}

func run() int {
	// Add pprof flag support
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to file")
	
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
		} else if argv[i] == "-v" {
			verboseMode = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--process-sequentially" {
			processSequentially = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--serial" {
			processSequentially = constants.ON
		} else if processSequentially == constants.OFF && argv[i] == "--sequential" {
			processSequentially = constants.ON
		} else if processFull == constants.OFF && argv[i] == "--process" {
			processFull = constants.ON
		} else if argv[i] == "--output-file" || argv[i] == "-o" {
			if i+1 < argc {
				outputFilename = argv[i+1]
			}
		} else if argv[i] == "--corpus-dir" || argv[i] == "-c" {
			if i+1 < argc {
				corpusDir = argv[i+1]
			}
		} else if argv[i] == "-cpuprofile" {
			if i+1 < argc {
				cpuprofile = argv[i+1]
			}
		} else if argv[i] == "-memprofile" {
			if i+1 < argc {
				memprofile = argv[i+1]
			}
		}
	}

	// Start CPU profiling if requested
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Call nlp() and return its exit code (line 60)
	exitCode := nlp.Nlp(verboseMode, processSequentially, processFull, outputFilename, corpusDir)
	
	// Write memory profile if requested
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	
	return exitCode
}
