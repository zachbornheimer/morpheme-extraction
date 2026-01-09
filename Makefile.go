# Makefile for Go version of morpheme-extraction
# Go 1.25+ required

.PHONY: all build clean test bench run install help

# Build variables
BINARY_NAME=morpheme-extract
CMD_DIR=./cmd/morpheme-extract
BUILD_DIR=.
GO=go

# Default target
all: build

# Build the binary
build:
	@echo "Building Go version..."
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with race detector (for development)
build-race:
	@echo "Building with race detector..."
	$(GO) build -race -o $(BUILD_DIR)/$(BINARY_NAME)-race $(CMD_DIR)

# Build optimized binary
build-optimized:
	@echo "Building optimized binary..."
	$(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built optimized: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

# Run with test corpus
run: build
	@echo "Running on test corpus..."
	./$(BINARY_NAME) --verbose --corpus-dir ./test-corpus/ --output-file go_output.txt

# Run in serial mode
run-serial: build
	@echo "Running in serial mode..."
	./$(BINARY_NAME) --verbose --serial --corpus-dir ./test-corpus/ --output-file go_output_serial.txt

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod tidy
	$(GO) mod download

# Clean build artifacts
clean:
	@echo "Cleaning Go build artifacts..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-race
	rm -f go_output.txt go_output_serial.txt
	rm -f coverage.out

# Run go fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

# Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/"; exit 1)
	golangci-lint run ./...

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install binary to GOPATH/bin
install:
	@echo "Installing to GOPATH/bin..."
	$(GO) install $(CMD_DIR)

# Help target
help:
	@echo "Morpheme Extraction - Go Version"
	@echo ""
	@echo "Available targets:"
	@echo "  make build           - Build the binary"
	@echo "  make build-race      - Build with race detector"
	@echo "  make build-optimized - Build optimized binary"
	@echo "  make test            - Run tests"
	@echo "  make bench           - Run benchmarks"
	@echo "  make run             - Run on test corpus"
	@echo "  make run-serial      - Run in serial mode"
	@echo "  make deps            - Install dependencies"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make fmt             - Format code"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run linters"
	@echo "  make coverage        - Generate coverage report"
	@echo "  make install         - Install to GOPATH/bin"
	@echo "  make help            - Show this help"

