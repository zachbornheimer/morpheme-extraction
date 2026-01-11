// main.go - Entry point for morpheme extraction system
// Line-for-line translation of nlp.c:main() lines 32-61
package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/zachbornheimer/morpheme-extraction/internal/config"
	"github.com/zachbornheimer/morpheme-extraction/internal/constants"
	"github.com/zachbornheimer/morpheme-extraction/internal/nlp"
)

// Profiling flags
var (
	cpuprofile string
	memprofile string
)

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize configuration
	cfg := config.NewConfig()
	cfg.OutputFilename = constants.ZedramOutput
	cfg.CorpusDir = constants.DefaultPath

	// Add pprof flag support
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to file")

	argc := len(os.Args)
	argv := os.Args

	// Parse command line arguments (lines 40-58)
	// Note: C used OpenMP parallel for which caused race conditions, we use serial
	for i := 0; i < argc; i++ {
		if cfg.VerboseMode == constants.OFF && argv[i] == "--verbose" {
			cfg.VerboseMode = constants.ON
		} else if argv[i] == "-v" {
			cfg.VerboseMode = constants.ON
		} else if cfg.ProcessSequentially == constants.OFF && argv[i] == "--process-sequentially" {
			cfg.ProcessSequentially = constants.ON
		} else if cfg.ProcessSequentially == constants.OFF && argv[i] == "--serial" {
			cfg.ProcessSequentially = constants.ON
		} else if cfg.ProcessSequentially == constants.OFF && argv[i] == "--sequential" {
			cfg.ProcessSequentially = constants.ON
		} else if cfg.ProcessFull == constants.OFF && argv[i] == "--process" {
			cfg.ProcessFull = constants.ON
		} else if argv[i] == "--output-file" || argv[i] == "-o" {
			if i+1 < argc {
				cfg.OutputFilename = argv[i+1]
			}
		} else if argv[i] == "--corpus-dir" || argv[i] == "-c" {
			if i+1 < argc {
				cfg.CorpusDir = argv[i+1]
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
	exitCode := nlp.Nlp(cfg)
	
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
