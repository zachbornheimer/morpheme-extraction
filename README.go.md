# Morpheme Extraction - Go Implementation

A language-agnostic, unsupervised morpheme extraction tool based on **Linguistic Fusion** theory (Bornheimer, 2013-2014).

This is a modern Go rewrite of the original C implementation, leveraging Go's concurrency, memory safety, and standard library.

## Overview

This tool automatically extracts morphemes (prefixes, suffixes, stems, infixes, and circumfixes) from any text corpus without prior knowledge of the language. It uses n-gram analysis and statistical pattern matching to identify morphological structure.

### Key Features

- **Language-agnostic**: Works with any human language
- **Unsupervised**: No training data or linguistic rules required
- **Parallel processing**: Uses goroutines for efficient n-gram similarity detection
- **Configurable thresholds**: Tune algorithm parameters for different languages/corpora
- **Memory-safe**: No manual memory management, no segfaults
- **Fast compilation**: Go compiles in seconds vs. minutes for C++

## What's New in the Go Version

### Improvements Over Original C Implementation

1. **Memory Safety**: No manual `malloc`/`free`, no memory leaks
2. **Concurrency**: Goroutines + channels instead of OpenMP
3. **Type Safety**: Compile-time type checking prevents many bugs
4. **Standard Library**: Built-in string manipulation, file I/O, testing
5. **Cross-platform**: Single binary works on Linux, macOS, Windows
6. **Modern Tooling**: `go test`, `go fmt`, `go vet`, benchmarking built-in

### Architecture

```
morpheme-extraction/
├── cmd/
│   └── morpheme-extract/       # Main CLI application
│       └── main.go
├── pkg/                        # Public packages
│   ├── config/                 # Configuration (thresholds, paths)
│   ├── corpus/                 # File reading & delimiter detection
│   ├── ngram/                  # N-gram building & similarity
│   └── morpheme/               # Morpheme extraction & classification
├── internal/
│   └── types/                  # Core data structures
└── test-corpus/                # Example corpus files
```

## Installation

### Prerequisites

- Go 1.25 or later
- GCC 15+ (for legacy C version only)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/zachbornheimer/morpheme-extraction.git
cd morpheme-extraction

# Build the Go version
make -f Makefile.go build

# Or build manually
go build -o morpheme-extract ./cmd/morpheme-extract
```

### Install to GOPATH

```bash
make -f Makefile.go install
# Binary now available as: morpheme-extract
```

## Usage

### Basic Usage

```bash
# Process all files in a directory
./morpheme-extract --corpus-dir ./test-corpus/ --output-file results.txt

# Enable verbose output
./morpheme-extract --verbose --corpus-dir ./my-corpus/

# Process files sequentially (one-by-one)
./morpheme-extract --serial --corpus-dir ./my-corpus/

# Process both ways (sequential + batch)
./morpheme-extract --process --corpus-dir ./my-corpus/
```

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--corpus-dir` | Directory containing corpus files | `./test-corpus/` |
| `--output-file` | Output file path | `languageModel.zdrm` |
| `--verbose` | Enable detailed logging | `false` |
| `--serial`, `--sequential` | Process files one-by-one | `false` |
| `--process` | Run both sequential and batch modes | `false` |

### Example Output

```
Found 1 file(s) in ./test-corpus/
Processing all files together...
Looking for word delimiter...
Found word delimiter: " "
Found 2211 words
Building n-grams...
Built 972 n-grams
Finding similar n-grams...
Found 241177 similar n-gram relationships
Extracting morphemes...
Extracted 661 morpheme candidates
Confirmed 381 morphemes

All Processed Files
===================
PREFIX: th
PREFIX: st
PREFIX: re
SUFFIX: ing
SUFFIX: ed
STEM: in
INFIX: tion
...
```

## Algorithm

The extraction process follows these steps:

### 1. Word Delimiter Detection
Automatically identifies the character(s) used to separate words (typically space, but works with any delimiter).

### 2. N-gram Construction
Builds n-grams (default size: 9) with context windows before and after each word:
```
Before: [word-4, word-3, word-2, word-1]
Target: word
After:  [word+1, word+2, word+3, word+4]
```

### 3. Similarity Detection
Finds n-grams in "parallel environments" (similar contexts). Two n-grams are similar if their combined context words overlap by at least 12.5% (configurable).

### 4. Morpheme Extraction
For each pair of similar n-grams, finds:
- **Longest common prefix** (potential prefix/stem)
- **Longest common suffix** (potential suffix/stem)
- **Internal matches** (potential infixes)

### 5. Classification
Classifies extracted morphemes as:
- **PREFIX**: Appears at beginning of words
- **SUFFIX**: Appears at end of words
- **STEM**: Appears standalone or in all positions
- **INFIX**: Appears word-internally
- **CIRCUMFIX**: Prefix+suffix pair that co-occurs (e.g., "en-...ed")

## Configuration

### Thresholds (from original research)

These values are based on the 2013-2014 research and match the original C implementation:

```go
NgramSize:              9     // N-gram window size
ThresholdSimilarNgrams: 12.5  // % overlap for "similar" n-grams
ThresholdCircumfix:     100.0 // % co-occurrence for circumfixes
ThresholdConfirmation:  10.0  // % frequency for morpheme confirmation
```

To modify these, edit `pkg/config/config.go` or create a custom `Config` struct.

## Development

### Running Tests

```bash
# Run all tests
make -f Makefile.go test

# Run with race detector
go test -race ./...

# Run specific package
go test ./pkg/morpheme/
```

### Benchmarking

```bash
make -f Makefile.go bench
```

### Code Quality

```bash
# Format code
make -f Makefile.go fmt

# Run static analysis
make -f Makefile.go vet

# Run linters (requires golangci-lint)
make -f Makefile.go lint
```

## Performance

The Go version is comparable in speed to the optimized C version:

- **War of the Worlds (12KB)**: ~0.5s
- **Large corpus (1MB)**: ~5-10s
- **Memory usage**: ~10-50MB depending on corpus size

Benefits of Go's goroutines:
- Scales to millions of concurrent goroutines (vs. limited OpenMP threads)
- Better CPU utilization on modern multi-core systems
- No manual tuning of thread count

## Comparison: C vs. Go

| Aspect | Original C (2013) | Go Rewrite (2025) |
|--------|-------------------|-------------------|
| **Lines of code** | ~1,800 | ~800 (cleaner abstractions) |
| **Memory safety** | Manual `malloc`/`free` | Garbage collected |
| **Concurrency** | OpenMP (`#pragma`) | Goroutines + channels |
| **Compilation** | ~10s with optimization | ~2s |
| **Cross-platform** | Makefile tweaks needed | Single `go build` |
| **Testing** | External framework needed | Built-in `go test` |
| **Dependencies** | GCC + OpenMP | Just Go toolchain |

## Research Background

This tool implements the algorithm described in:

- **Bornheimer, Z.** (2013). *Nonstatistical Language-Blind Morpheme Extraction*. Fall 2013 Research Summary, University of South Florida.
- **Bornheimer, Z.** (2014). *Linguistic Fusion: A Unified Theory of Language Acquisition*. Spring 2014 Research Paper, USF Honors College.

### Core Insight

> "Acquisition and production are governed by rules whose features are visible in the output."

This means you can reverse-engineer language structure by finding patterns in text, without knowing the language beforehand.

## Applications

1. **Low-resource languages**: Extract morphology when linguists unavailable
2. **Computational linguistics**: Bootstrap morphological analyzers
3. **Language documentation**: Assist field linguists in endangered language work
4. **Code tokenization**: Apply to programming languages for better code search
5. **Hybrid NLP**: Combine rule-based morphology with neural approaches

## Future Work

### Planned Features

- [ ] JSON/protobuf output format
- [ ] HTTP API for integration
- [ ] Streaming for very large corpora
- [ ] Regex pattern generation (from original C code)
- [ ] Lexical categorization (Phase 2 of Linguistic Fusion)
- [ ] Syntax modeling (Phase 3)
- [ ] Semantic mapping (Phase 4)

### Performance Optimizations

- [ ] Bloom filters for n-gram similarity pre-filtering
- [ ] Trie data structure for morpheme storage
- [ ] Incremental processing for streaming inputs
- [ ] Memory pooling for large corpora

## License

GPLv2 (matching original C implementation)

## Citation

If you use this tool in research, please cite:

```bibtex
@software{bornheimer2025morpheme,
  author = {Bornheimer, Zachary},
  title = {Morpheme Extraction: A Language-Agnostic Unsupervised Approach},
  year = {2025},
  url = {https://github.com/zachbornheimer/morpheme-extraction},
  note = {Go rewrite of 2013 C implementation}
}
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run `make test` and `make lint`
5. Submit a pull request

## Contact

For questions about the algorithm or implementation, please open a GitHub issue or contact zachary@bornheimer.com.

---

**Note**: This Go version was created in January 2025 as a modernization of the original 2013 C implementation. The algorithm and thresholds remain faithful to the original research while leveraging modern language features.

