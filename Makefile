# Morpheme Extraction Build Configuration

.PHONY: all build build-standard build-greentea clean test profile

# Default: Build with Green Tea GC (recommended, 4.6% faster)
all: build-greentea

# Build with Green Tea GC (Go 1.25+ experimental GC)
build-greentea:
	GOEXPERIMENT=greenteagc go build -o nlp-go ./cmd/morpheme-extract
	@echo "Built with Green Tea GC (4.6% faster)"

# Build with standard GC (fallback for Go < 1.25)
build-standard:
	go build -o nlp-go ./cmd/morpheme-extract
	@echo "Built with standard GC"

# Alias for default
build: build-greentea

# Clean build artifacts
clean:
	rm -f nlp-go nlp-go-greentea nlp-go-test
	rm -f *.prof languageModel.zy

# Run tests
test:
	go test ./...

# Profile with Green Tea GC
profile:
	@echo "Running with CPU and memory profiling..."
	GOEXPERIMENT=greenteagc go build -o nlp-go-test ./cmd/morpheme-extract
	rm -f cpu.prof mem.prof languageModel.zy
	./nlp-go-test -cpuprofile cpu.prof -memprofile mem.prof test-corpus/
	@echo ""
	@echo "=== CPU Profile Top 10 ==="
	go tool pprof -top -nodecount=10 cpu.prof
	@echo ""
	@echo "=== Memory Profile Top 10 ==="
	go tool pprof -top -nodecount=10 mem.prof

# Profile without GC (for baseline comparison)
profile-nogc:
	@echo "Running with GC disabled (baseline)..."
	GOEXPERIMENT=greenteagc go build -o nlp-go-test ./cmd/morpheme-extract
	rm -f cpu-nogc.prof mem-nogc.prof languageModel.zy
	GOGC=off ./nlp-go-test -cpuprofile cpu-nogc.prof -memprofile mem-nogc.prof test-corpus/
	@echo ""
	@echo "=== CPU Profile Top 10 (No GC) ==="
	go tool pprof -top -nodecount=10 cpu-nogc.prof

# Run and show results
run:
	./nlp-go test-corpus/
	@echo ""
	@echo "=== Results ==="
	@echo "Total morphemes: $$(wc -l < languageModel.zy)"
	@echo "Unique STEMs: $$(grep -a '^STEM:' languageModel.zy | sort -u | wc -l)"
	@echo "INFIXes: $$(grep -ac '^INFIX:' languageModel.zy)"
	@echo "PREFIXes: $$(grep -ac '^PREFIX:' languageModel.zy)"
	@echo "SUFFIXes: $$(grep -ac '^SUFFIX:' languageModel.zy)"
