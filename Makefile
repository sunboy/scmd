# scmd Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/scmd/scmd/pkg/version.Version=$(VERSION) \
	-X github.com/scmd/scmd/pkg/version.Commit=$(COMMIT) \
	-X github.com/scmd/scmd/pkg/version.Date=$(DATE)

.PHONY: all build test lint clean install dev fmt vet coverage deps help

# Default target
all: lint test build

# Build the binary
build:
	@echo "Building scmd..."
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/scmd ./cmd/scmd

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-arm64 ./cmd/scmd
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-amd64 ./cmd/scmd
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-amd64 ./cmd/scmd
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-arm64 ./cmd/scmd
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-windows-amd64.exe ./cmd/scmd

# Run tests
test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.out ./...

# Run tests with short flag (for CI)
test-short:
	@echo "Running short tests..."
	go test -short -race ./...

# Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	go test -race -v ./...

# Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet only"; \
		go vet ./...; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Run in development mode
dev:
	@echo "Running scmd..."
	go run ./cmd/scmd

# Install to /usr/local/bin
install: build
	@echo "Installing scmd..."
	cp bin/scmd /usr/local/bin/
	@echo "Installed to /usr/local/bin/scmd"

# Install to GOPATH/bin
install-go: build
	@echo "Installing scmd to GOPATH..."
	go install ./cmd/scmd

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/ coverage.out coverage.html

# Tidy dependencies
deps:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# Generate (placeholder for future code generation)
generate:
	@echo "Running go generate..."
	go generate ./...

# Check for outdated dependencies
outdated:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Security audit
audit:
	@echo "Running security audit..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Help
help:
	@echo "scmd Makefile targets:"
	@echo "  all        - Run lint, test, and build (default)"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for all platforms"
	@echo "  test       - Run all tests with coverage"
	@echo "  test-short - Run short tests"
	@echo "  test-v     - Run tests with verbose output"
	@echo "  coverage   - Generate coverage HTML report"
	@echo "  lint       - Run linters"
	@echo "  vet        - Run go vet"
	@echo "  fmt        - Format code"
	@echo "  dev        - Run in development mode"
	@echo "  install    - Install to /usr/local/bin"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Tidy and verify dependencies"
	@echo "  help       - Show this help message"
